package config

import (
	"log"
	//libraries
	"github.com/caarlos0/env/v6" //for parsing the env file
	"github.com/joho/godotenv"   //To load up the .env file at start
)

// variable of type AppConfig struct
// declaring the gloabal variable so it can be used all over the project
var Cfg AppConfig

type AppConfig struct {
	Server struct {
		APIBaseURL string `env:"API_BASE_URL"`
		// Use a string for PORT, and provide a default for easy local dev
		PORT string `env:"PORT" envDefault:"8000"`
	}

	// EmailConfig holds all settings for the email service.
	Email struct {
		SMTPHost    string `env:"SMTP_HOST,required"`
		SMTPPort    string `env:"SMTP_PORT,required"`
		SMTPUser    string `env:"SMTP_USER,required"`
		SMTPPass    string `env:"SMTP_PASS,required"`
		EmailFrom   string `env:"EMAIL_FROM,required"`
		ResendApi   string `env:"RESEND_API"`
		SendGridKey string `env:"SEND_GRID_KEY,required"`
	}

	// DatabaseConfig holds the database connection string.
	Database struct {
		URL string `env:"DATABASE_URL,required"`
	}

	// TokenConfig holds the secret for JWT/JWE tokens.
	Secret struct {
		TOKEN string `env:"TOKEN_KEY,required"`
	}

	CHECK_ENV struct {
		ENV string `env:"APP_ENV"`
	}

	OAuth struct {
		ClientID     string `env:"GOOGLE_OAUTH_CLIENT_ID"`
		ClientSecret string `env:"GOOGLE_OAUTH_CLIENT_SECRET"`
	}
	MinIO struct {
		Endpoint string `env:"MINIO_ENDPOINT,required"`
		Username string `env:"MINIO_ROOT_USER,required"`
		Password string `env:"MINIO_ROOT_PASSWORD,required"`
		Bucket   string `env:"MINIO_BUCKET,required"`
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
