package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const defaultInsurancePrompt = "Eres un vendedor de seguros profesional, amable y servicial. Tu objetivo es asesorar a los clientes sobre seguros de vida, hogar, autos y salud, respondiendo de manera concisa, clara y empática. Responde siempre en español y mantén tus mensajes amigables y formateados para WhatsApp (usa emojis, párrafos cortos y viñetas cuando sea apropiado). Si te preguntan cosas sin relación con seguros aclara que no eres un experto en el tema y redirige la pregunta hacia seguros de vida, hogar, autos y salud."

// Config holds all configuration details for our server and WhatsApp connection.
type Config struct {
	Port               string
	VerifyToken        string
	AccessToken        string
	PhoneNumberID      string
	APIVersion         string
	GeminiAPIKey       string
	GeminiModel        string
	GeminiSystemPrompt string
}

// LoadConfig loads variables from .env (if it exists) and returns the populated Config.
func LoadConfig() (*Config, error) {
	// Try loading from .env if present. Ignored if file doesn't exist.
	_ = loadEnvFile(".env")

	cfg := &Config{
		Port:               getEnv("PORT", "8093"),
		VerifyToken:        os.Getenv("WHATSAPP_VERIFY_TOKEN"),
		AccessToken:        os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		PhoneNumberID:      os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		APIVersion:         getEnv("WHATSAPP_API_VERSION", "v20.0"),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		GeminiModel:        getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		GeminiSystemPrompt: getEnv("GEMINI_SYSTEM_PROMPT", defaultInsurancePrompt),
	}

	// Validate required variables. We only return error for required variables.
	if cfg.VerifyToken == "" {
		return nil, fmt.Errorf("missing WHATSAPP_VERIFY_TOKEN in environment")
	}
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("missing WHATSAPP_ACCESS_TOKEN in environment")
	}
	if cfg.PhoneNumberID == "" {
		return nil, fmt.Errorf("missing WHATSAPP_PHONE_NUMBER_ID in environment")
	}
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("missing GEMINI_API_KEY in environment")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// loadEnvFile parses key-value pairs from a file and sets them as env vars if not already defined.
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip quotes if any
		val = strings.Trim(val, `"'`)

		// Set key only if it doesn't already exist in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
