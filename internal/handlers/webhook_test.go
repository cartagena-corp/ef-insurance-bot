package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ef-insurance-bot/models"
)

// mockSender implements MessageSender for testing.
type mockSender struct{}

func (m *mockSender) SendTextMessage(to, text string) (*models.SendMessageResponse, error) {
	return &models.SendMessageResponse{}, nil
}

// mockResponder implements ResponseGenerator for testing.
type mockResponder struct{}

func (m *mockResponder) GenerateResponse(senderID, userPrompt string) (string, error) {
	return "mock response", nil
}

func TestVerifyWebhook_Success(t *testing.T) {
	handler := NewWebhookHandler("supersecret", &mockSender{}, &mockResponder{})

	req, err := http.NewRequest("GET", "/webhook?hub.mode=subscribe&hub.verify_token=supersecret&hub.challenge=test_challenge", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBody := "test_challenge"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}

func TestVerifyWebhook_Forbidden(t *testing.T) {
	handler := NewWebhookHandler("supersecret", &mockSender{}, &mockResponder{})

	req, err := http.NewRequest("GET", "/webhook?hub.mode=subscribe&hub.verify_token=wrongtoken&hub.challenge=test_challenge", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}

func TestHandleWebhook_MethodNotAllowed(t *testing.T) {
	handler := NewWebhookHandler("supersecret", &mockSender{}, &mockResponder{})

	req, err := http.NewRequest("DELETE", "/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleWebhook(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
