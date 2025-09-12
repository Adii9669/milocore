package config

import (
	"fmt"
	"os"
)

// AppConfig holds the application's configuration in a structured way.
type AppConfig struct {
	Server   ServerConfig
	Email    EmailConfig
	Database DatabaseConfig
	Token    TokenConfig
}

// ServerConfig holds server-specific settings.
type ServerConfig struct {
	APIBaseURL string
}

// EmailConfig holds all settings for the email service.
type EmailConfig struct {
	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
	EmailFrom string
}

// DatabaseConfig holds the database connection string.
type DatabaseConfig struct {
	URL string
}

// TokenConfig holds the secret for JWT/JWE tokens.
type TokenConfig struct {
	Secret string
}

// Cfg is the global, accessible configuration instance.
var Cfg AppConfig

// LoadConfig reads environment variables and populates the Cfg struct.
// It should be called once at application startup.
func LoadConfig() error {
	getEnv := func(key string) (string, error) {
		value := os.Getenv(key)
		if value == "" {
			return "", fmt.Errorf("environment variable %s not set", key)
		}
		return value, nil
	}

	var err error

	// --- Load Server Config ---
	if Cfg.Server.APIBaseURL, err = getEnv("API_BASE_URL"); err != nil {
		return err
	}

	// --- Load Database Config ---
	if Cfg.Database.URL, err = getEnv("DATABASE_URL"); err != nil {
		return err
	}

	// --- Load Token Config ---
	if Cfg.Token.Secret, err = getEnv("TOKEN_KEY"); err != nil {
		return err
	}

	// --- Load Email Config ---
	if Cfg.Email.SMTPHost, err = getEnv("SMTP_HOST"); err != nil {
		return err
	}
	if Cfg.Email.SMTPPort, err = getEnv("SMTP_PORT"); err != nil {
		return err
	}
	if Cfg.Email.SMTPUser, err = getEnv("SMTP_USER"); err != nil {
		return err
	}
	if Cfg.Email.SMTPPass, err = getEnv("SMTP_PASS"); err != nil {
		return err
	}
	if Cfg.Email.EmailFrom, err = getEnv("EMAIL_FROM"); err != nil {
		return err
	}

	return nil
}
