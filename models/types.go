package models

// WebhookRequest represents the incoming webhook payload from Meta when an event occurs.
type WebhookRequest struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// WebhookEntry represents a change event inside the Webhook payload.
type WebhookEntry struct {
	ID      string          `json:"id"`
	Changes []WebhookChange `json:"changes"`
}

// WebhookChange contains the fields that have updated.
type WebhookChange struct {
	Value WebhookValue `json:"value"`
	Field string       `json:"field"`
}

// WebhookValue contains details about the incoming messages, contacts, and metadata.
type WebhookValue struct {
	MessagingProduct string           `json:"messaging_product"`
	Metadata         WebhookMetadata  `json:"metadata"`
	Contacts         []WebhookContact `json:"contacts"`
	Messages         []WebhookMessage `json:"messages"`
}

// WebhookMetadata lists metadata about the receiver.
type WebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// WebhookContact describes the profile and ID of the sender.
type WebhookContact struct {
	Profile Profile `json:"profile"`
	WaID    string  `json:"wa_id"` // WhatsApp ID, typically the phone number.
}

// Profile holds the profile information.
type Profile struct {
	Name string `json:"name"`
}

// WebhookMessage represents an incoming message.
type WebhookMessage struct {
	From      string             `json:"from"` // Sender's phone number
	ID        string             `json:"id"`   // Unique message ID
	Timestamp string             `json:"timestamp"`
	Type      string             `json:"type"` // e.g. "text", "image", "location", "document", etc.
	Text      *TextMessageDetail `json:"text,omitempty"`
}

// TextMessageDetail contains the actual body text of the text message.
type TextMessageDetail struct {
	Body string `json:"body"`
}

// SendMessageRequest represents the POST payload to send a text message via Meta Graph API.
type SendMessageRequest struct {
	MessagingProduct string           `json:"messaging_product"`
	RecipientType    string           `json:"recipient_type"`
	To               string           `json:"to"`
	Type             string           `json:"type"`
	Text             *TextMessageSend `json:"text,omitempty"`
}

// TextMessageSend details the text payload when sending a message.
type TextMessageSend struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

// SendMessageResponse represents the payload returned by Meta on success.
type SendMessageResponse struct {
	MessagingProduct string                  `json:"messaging_product"`
	Contacts         []SendMessageContact    `json:"contacts"`
	Messages         []SendMessageStatusInfo `json:"messages"`
}

// SendMessageContact maps the recipient info in the response.
type SendMessageContact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

// SendMessageStatusInfo details the message ID of the sent message.
type SendMessageStatusInfo struct {
	ID string `json:"id"`
}
