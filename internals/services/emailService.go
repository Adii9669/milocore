package services

import (
	"chat-server/internals/config"
	"fmt"
	"log"

	"github.com/resend/resend-go/v3"
)

type EmailService struct {
	apiKey string
	from   string
	client *resend.Client
}

func NewEmailService() *EmailService {
	return &EmailService{
		client: resend.NewClient(config.Cfg.Email.ResendApi),
		apiKey: config.Cfg.Email.ResendApi,
		from:   config.Cfg.Email.EmailFrom,
	}
}

func (e *EmailService) SendOTP(toEmail string, otp string) error {

	params := &resend.SendEmailRequest{
		From:    e.from,
		To:      []string{toEmail},
		Subject: "Verify Your Account",
		Html: fmt.Sprintf(`
			<h2>Verify Your Account</h2>
			<h1>%s</h1>
			<p>This OTP expires in 5 minutes.</p>
		`, otp),
	}

	_, err := e.client.Emails.Send(params)
	if err != nil {
		log.Println("RESEND ERROR:", err)
		return err
	}

	log.Println("📨 SendOTP called")
	return nil
}
