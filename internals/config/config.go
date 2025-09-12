package config

import (
	"log"
	//libraries
	"github.com/caarlos0/env/v6" //for parsing the env file
	"github.com/joho/godotenv"   //To load up the .env file at start
)

var Cfg AppConfig

type AppConfig struct {
	Server struct {
		APIBaseURL string `env:"API_BASE_URL"`
		// Use a string for PORT, and provide a default for easy local dev
		PORT string `env:"PORT" envDefault:"8000"`
	}

	// EmailConfig holds all settings for the email service.
	Email struct {
		SMTPHost  string `env:"SMTP_HOST,required"`
		SMTPPort  string `env:"SMTP_PORT,required"`
		SMTPUser  string `env:"SMTP_USER,required"`
		SMTPPass  string `env:"SMTP_PASS,required"`
		EmailFrom string `env:"EMAIL_FROM,required"`
	}

	// DatabaseConfig holds the database connection string.
	Database struct {
		URL string `env:"DATABASE_URL,required"`
	}

	// TokenConfig holds the secret for JWT/JWE tokens.
	Token struct {
		Secret string `env:"TOKEN_KEY,required"`
	}
}

func LoadConfig() error {

	//Load the .env file here
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file FOUND, reading env")
	}

	//parse env in the strcut i created
	if err := env.Parse(&Cfg); err != nil {
		return err
	}
	return nil
}
