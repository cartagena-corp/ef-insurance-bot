package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"ef-insurance-bot/models"
)

// MessageSender defines the contract for sending messages to a user.
type MessageSender interface {
	SendTextMessage(to, text string) (*models.SendMessageResponse, error)
}

// ResponseGenerator defines the contract for generating AI responses.
type ResponseGenerator interface {
	GenerateResponse(senderID, userPrompt string) (string, error)
}

// WebhookHandler contains the verify token, message sender and AI response generator.
type WebhookHandler struct {
	verifyToken string
	sender      MessageSender
	responder   ResponseGenerator
}

// NewWebhookHandler initializes the handler with its dependencies.
func NewWebhookHandler(verifyToken string, sender MessageSender, responder ResponseGenerator) *WebhookHandler {
	return &WebhookHandler{
		verifyToken: verifyToken,
		sender:      sender,
		responder:   responder,
	}
}

// RegisterRoutes registers the handlers onto the provided server mux.
func (h *WebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/webhook", h.HandleWebhook)
}

// HandleWebhook acts as the central router for webhook GET verification and POST payloads.
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.verifyWebhook(w, r)
	case http.MethodPost:
		h.receiveMessage(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// verifyWebhook processes Meta's verification request (GET /webhook).
func (h *WebhookHandler) verifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == h.verifyToken {
		log.Println("Webhook verified successfully!")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	log.Printf("Webhook verification failed. Mode: %s, Token: %s", mode, token)
	w.WriteHeader(http.StatusForbidden)
}

// receiveMessage processes incoming events (POST /webhook).
func (h *WebhookHandler) receiveMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload models.WebhookRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error unmarshaling request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Always reply with 200 OK immediately to acknowledge Meta webhook.
	// This prevents Meta from marking our webhook as failed or retrying.
	w.WriteHeader(http.StatusOK)

	// Process message in a separate goroutine to avoid holding the webhook request open.
	go h.processPayload(payload)
}

func (h *WebhookHandler) processPayload(payload models.WebhookRequest) {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			// Skip status updates (like delivered, read) to avoid infinite loops or unnecessary errors.
			if len(change.Value.Messages) == 0 {
				continue
			}

			for _, msg := range change.Value.Messages {
				// Only process incoming text messages
				if msg.Type != "text" || msg.Text == nil {
					log.Printf("Received non-text message of type: %s", msg.Type)
					continue
				}

				senderPhoneNumber := msg.From
				messageBody := strings.TrimSpace(msg.Text.Body)

				log.Printf("Incoming message from %s: %s", senderPhoneNumber, messageBody)

				// Generate bot reply text using Gemini API
				replyText, err := h.responder.GenerateResponse(senderPhoneNumber, messageBody)
				if err != nil {
					log.Printf("Failed to generate response from Gemini: %v", err)
					replyText = "Disculpa, he tenido un inconveniente al procesar tu solicitud. Por favor, intenta de nuevo en unos momentos."
				}

				// Send the reply back to the user
				resp, err := h.sender.SendTextMessage(senderPhoneNumber, replyText)
				if err != nil {
					log.Printf("Failed to send reply to %s: %v", senderPhoneNumber, err)
				} else if len(resp.Messages) > 0 {
					log.Printf("Successfully sent reply to %s. Message ID: %s", senderPhoneNumber, resp.Messages[0].ID)
				} else {
					log.Printf("Successfully sent reply to %s", senderPhoneNumber)
				}
			}
		}
	}
}
