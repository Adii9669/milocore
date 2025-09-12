package utils

import (
	"chat-server/internals/config"
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendVerificationEmail(to, url string) error {

	config := config.Cfg

	smtpHost := config.Email.SMTPHost
	smtpPortStr := config.Email.SMTPPort
	smtpUser := config.Email.SMTPUser
	smtpPass := config.Email.SMTPPass
	emailFrom := config.Email.EmailFrom

	if smtpHost == "" {
		return fmt.Errorf("host")
	}

	if smtpUser == "" {
		return fmt.Errorf("user")
	}

	if smtpPortStr == "" {
		return fmt.Errorf("port")
	}

	if emailFrom == "" {
		return fmt.Errorf("emai")
	}

	// if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" || emailFrom == "" {
	// 	return fmt.Errorf("SMTP configuration is incomplete. Please check environment variables")
	// }

	//changin the port from string to int
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP Port: %w", err)
	}

	//now using the mail
	m := gomail.NewMessage()

	m.SetHeader("From", emailFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Welcome! To MiloVerce Please Verify you account")

	emailBody := fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; line-height: 1.6;">
        <h2>Welcome to Go Chat!</h2>
        <p>Thanks for registering. Please click the button below to verify your account and get started:</p>
        <p style="margin: 20px 0;">
            <a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #007bff; color: #ffffff; text-decoration: none; border-radius: 5px;">
                Verify My Account
            </a>
        </p>
        <p>If you did not register for an account, please disregard this email.</p>
        <p>Thanks,<br>The Go Chat Team</p>
    </div>`, url)

	m.SetBody("text/html", emailBody)

	//sending email by smtp
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("\nFailed To Send the  mail: %w:\n", err)
	}

	return nil
}
