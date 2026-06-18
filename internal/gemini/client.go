package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"ef-insurance-bot/config"
)

// Client interacts with Google's Gemini API and tracks chat sessions.
type Client struct {
	apiKey       string
	model        string
	systemPrompt string
	httpClient   *http.Client

	// Thread-safe session memory
	mu       sync.Mutex
	sessions map[string][]Content
}

// NewClient initializes a Gemini client with configurations.
// The model version and system prompt are loaded from environment variables
// via config.Config (GEMINI_MODEL and GEMINI_SYSTEM_PROMPT).
func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiKey:       cfg.GeminiAPIKey,
		model:        cfg.GeminiModel,
		systemPrompt: cfg.GeminiSystemPrompt,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		sessions: make(map[string][]Content),
	}
}

// Request represents the JSON request body for Gemini API.
type Request struct {
	SystemInstruction *Instruction `json:"systemInstruction,omitempty"`
	Contents          []Content    `json:"contents"`
}

type Instruction struct {
	Parts []Part `json:"parts"`
}

type Content struct {
	Role  string `json:"role"` // "user" or "model"
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// Response represents the JSON response body from Gemini API.
type Response struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content CandidateContent `json:"content"`
}

type CandidateContent struct {
	Parts []Part `json:"parts"`
}

const maxHistorySize = 20

// GenerateResponse sends the prompt to Gemini along with the session's chat history.
func (c *Client) GenerateResponse(senderID string, userPrompt string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Retrieve the existing conversation history for this sender
	history := c.sessions[senderID]

	// 2. Prepare the new user message
	userMsg := Content{
		Role: "user",
		Parts: []Part{
			{Text: userPrompt},
		},
	}

	// 3. Temporarily append user's new message to construct the request contents
	reqContents := append(history, userMsg)

	// 4. Dispatch the API call
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)

	reqPayload := Request{
		SystemInstruction: &Instruction{
			Parts: []Part{
				{Text: c.systemPrompt},
			},
		},
		Contents: reqContents,
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Gemini request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var geminiResp Response
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response received from Gemini")
	}

	replyText := geminiResp.Candidates[0].Content.Parts[0].Text

	// 5. Build and save the updated history (only on API call success)
	modelMsg := Content{
		Role: "model",
		Parts: []Part{
			{Text: replyText},
		},
	}

	// Save permanently
	updatedHistory := append(history, userMsg, modelMsg)
	c.sessions[senderID] = trimHistory(updatedHistory, maxHistorySize)

	return replyText, nil
}

// trimHistory slices history to keep at most N messages.
// It removes oldest pairs to ensure the history starts with a "user" message and alternates correctly.
func trimHistory(history []Content, max int) []Content {
	if len(history) <= max {
		return history
	}

	toRemove := len(history) - max
	// Ensure we remove an even number of elements so the remaining history
	// still alternates properly (user -> model -> user -> model) and starts with a user turn.
	if toRemove%2 != 0 {
		toRemove++
	}

	if toRemove >= len(history) {
		return []Content{}
	}

	return history[toRemove:]
}
