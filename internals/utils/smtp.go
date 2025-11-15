package utils

import (
	"net/smtp"
	"os"
)

func SendMai(to string, subject string, body string) error {

	from := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	return smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))

}
