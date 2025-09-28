package auth

import (
	"chat-server/internals/config"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"gopkg.in/gomail.v2"
)

func generateOtp() (string, error) {
	const otpLength = 6
	var otp string
	for i := 0; i < otpLength; i++ {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			// Handle the error, as rand.Int can return one
			return "", err
		}

		otp += digit.String()
	}
	return otp, nil
}

func SendOTP(toEmail string) (string, error) {

	otp, err := generateOtp()
	if err != nil {
		return "", fmt.Errorf("failed to generateOtp: %w", err)
	}

	//load the email from config
	cfg := config.Cfg.Email
	port, _ := strconv.Atoi(cfg.SMTPPort)

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.EmailFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Verify Your Account")
	m.SetBody("text/html", fmt.Sprintf("Hello,<br><br>Your verification code is: <strong>%s</strong>", otp))

	// 4. Set up the SMTP dialer and send the email
	d := gomail.NewDialer(cfg.SMTPHost, port, cfg.SMTPUser, cfg.SMTPPass)

	// send the email
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send verification email to %s: %v", toEmail, err)
		return "", err
	}

	log.Printf("Successfully sent verification OTP to %s", toEmail)
	// 5. Return the generated OTP so it can be saved to the database
	return otp, nil
}
