package repository

import (
	"fmt"
	"net/smtp"

	"github.com/gianlucafrei/GoAM/internal/logger"
)

type DefaultEmailSender struct {
}

func NewDefaultEmailSender() EmailSender {
	return &DefaultEmailSender{}
}

func (s *DefaultEmailSender) SendEmail(subject, body, recipientEmail, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail string) error {
	// Create a channel to receive the error
	errChan := make(chan error, 1)

	// Start the email sending in a goroutine
	go func() {
		// Connect to the remote SMTP server.
		smtpServerString := fmt.Sprintf("%s:%s", smtpServer, smtpPort)
		auth := smtp.PlainAuth("", smtpSenderEmail, smtpPassword, smtpServer)

		err := smtp.SendMail(smtpServerString, auth, smtpSenderEmail, []string{recipientEmail}, []byte(body))
		errChan <- err

		logger.InfoWithFieldsNoContext("Email sent", map[string]interface{}{
			"subject":      subject,
			"recipient":    recipientEmail,
			"smtpServer":   smtpServer,
			"smtpPort":     smtpPort,
			"smtpUsername": smtpUsername,
		})
	}()

	// Return immediately, the email will be sent in the background
	return nil
}
