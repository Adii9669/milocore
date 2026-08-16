package services

import (
	"chat-server/internals/config"
	"fmt"
	"log"

	"github.com/resend/resend-go/v3"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
	apiKey      string
	sendGridKey string
	from        string
	client      *resend.Client
}

func NewEmailService() *EmailService {
	return &EmailService{
		client:      resend.NewClient(config.Cfg.Email.ResendApi),
		apiKey:      config.Cfg.Email.ResendApi,
		from:        config.Cfg.Email.EmailFrom,
		sendGridKey: config.Cfg.Email.SendGridKey,
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

func (e *EmailService) SendOTPWITHSMTP(toEmail string, otp string) error {

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

func (e *EmailService) SendOTPWithGrid(toEmail string, otp string) error {
	from := mail.NewEmail("Miloverse", e.from)
	to := mail.NewEmail("", toEmail) // ← dynamic recipient
	subject := "Verify Your Account"
	plainText := fmt.Sprintf("Your verification code is: %s\nExpires in 5 minutes.", otp)
	htmlContent := fmt.Sprintf(`
        <h2>Verify Your Account</h2>
        <h1 style="letter-spacing: 8px;">%s</h1>
        <p>This code expires in 5 minutes.</p>
    `, otp)

	message := mail.NewSingleEmail(from, subject, to, plainText, htmlContent)
	client := sendgrid.NewSendClient(e.sendGridKey)

	response, err := client.Send(message)
	if err != nil {
		log.Println("SENDGRID ERROR:", err)
		return err
	}

	if response.StatusCode >= 400 {
		log.Printf("SENDGRID bad status: %d %s", response.StatusCode, response.Body)
		return fmt.Errorf("sendgrid failed with status %d", response.StatusCode)
	}

	log.Println("📨 OTP sent via SendGrid to:", toEmail)
	return nil
}
