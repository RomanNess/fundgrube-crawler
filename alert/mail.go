package alert

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendAlertMailBytes(subject string, content []byte) error {
	smtpHost := env("SMTP_SERVER", "smtp.gmail.com")
	smtpPort := env("SMTP_PORT", "587")
	username := env("SMTP_USERNAME", "n/a")
	password := env("SMTP_PASSWORD", "n/a")
	recipients := []string{env("SMTP_RECIPIENT", "n/a")}

	auth := smtp.PlainAuth("", username, password, smtpHost)
	messageBytes := append([]byte(fmt.Sprintf("Subject: %s\n\n", subject)), content...)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, username, recipients, messageBytes)
}

func SendAlertMail(subject string, content string) error {
	return SendAlertMailBytes(subject, []byte(content))
}

func env(key string, defaultValue string) string {
	value, present := os.LookupEnv(key)
	if present {
		return value
	}
	return defaultValue
}
