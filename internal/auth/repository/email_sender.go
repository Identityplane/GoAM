package repository

import (
	"fmt"
	"net/smtp"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
)

type DefaultEmailSender struct {
}

func NewDefaultEmailSender() model.EmailSender {
	return &DefaultEmailSender{}
}

func (s *DefaultEmailSender) SendEmail(subject, body, recipientEmail, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail string) error {
	log := logger.GetGoamLogger()

	// Create a channel to receive the error
	errChan := make(chan error, 1)

	// Start the email sending in a goroutine
	go func() {
		// Connect to the remote SMTP server.
		smtpServerString := fmt.Sprintf("%s:%s", smtpServer, smtpPort)
		auth := smtp.PlainAuth("", smtpSenderEmail, smtpPassword, smtpServer)

		err := smtp.SendMail(smtpServerString, auth, smtpSenderEmail, []string{recipientEmail}, []byte(body))
		errChan <- err

		log.Info().
			Str("to", recipientEmail).
			Str("subject", subject).
			Msg("email sent")
	}()

	// Return immediately, the email will be sent in the background
	return nil
}
