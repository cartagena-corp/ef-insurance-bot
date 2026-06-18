package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ef-insurance-bot/config"
	"ef-insurance-bot/models"
)

// Client wraps the WhatsApp Cloud API endpoints and authentication keys.
type Client struct {
	accessToken   string
	phoneNumberID string
	apiVersion    string
	httpClient    *http.Client
}

// NewClient initializes a WhatsApp client with the provided configurations.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		accessToken:   cfg.AccessToken,
		phoneNumberID: cfg.PhoneNumberID,
		apiVersion:    cfg.APIVersion,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendTextMessage sends a text response to the specified WhatsApp ID (phone number).
func (c *Client) SendTextMessage(to string, text string) (*models.SendMessageResponse, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, c.phoneNumberID)

	reqPayload := models.SendMessageRequest{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: &models.TextMessageSend{
			PreviewURL: false,
			Body:       text,
		},
	}

	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request to meta: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch request to meta: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("meta api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var sendResp models.SendMessageResponse
	if err := json.Unmarshal(respBody, &sendResp); err != nil {
		return nil, fmt.Errorf("failed to parse message response: %w", err)
	}

	return &sendResp, nil
}
