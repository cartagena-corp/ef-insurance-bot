package main

import (
	"fmt"
	"log"
	"net/http"

	"ef-insurance-bot/config"
	"ef-insurance-bot/internal/gemini"
	"ef-insurance-bot/internal/handlers"
	"ef-insurance-bot/internal/whatsapp"
)

func main() {
	log.Println("Starting WhatsApp Chatbot service...")

	// 1. Load configuration (from environment and .env)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Printf("Configuration loaded successfully. Webhook port: %s", cfg.Port)
	log.Printf("Gemini model: %s", cfg.GeminiModel)

	// 2. Initialize WhatsApp API client
	waClient := whatsapp.NewClient(cfg)

	// 3. Initialize Gemini AI client
	aiClient := gemini.NewClient(cfg)

	// 4. Initialize Webhook handler (using interfaces for loose coupling)
	webhook := handlers.NewWebhookHandler(cfg.VerifyToken, waClient, aiClient)

	// 5. Register routing
	mux := http.NewServeMux()
	webhook.RegisterRoutes(mux)

	// Health check and root route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("WhatsApp Chatbot Go service is running! 🚀"))
	})

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server is starting to listen on %s", serverAddr)

	// 6. Start HTTP server
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
