package repository

import (
	"context"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	services "github.com/Identityplane/GoAM/pkg/services"
)

type EmailSenderImpl struct {
	tenant       string
	realm        string
	emailService services.EmailService
}

func NewEmailSender(tenant, realm string, emailService services.EmailService) model.EmailSender {
	return &EmailSenderImpl{
		tenant:       tenant,
		realm:        realm,
		emailService: emailService,
	}
}

func (e *EmailSenderImpl) SendEmail(email *model.SendEmailParams) error {
	// Fire-and-forget to keep request latency low.
	go func() {
		log := logger.GetGoamLogger()
		if err := e.emailService.SendEmail(e.tenant, e.realm, email); err != nil {
			log.Error().
				Err(err).
				Str("tenant", e.tenant).
				Str("realm", e.realm).
				Msg("async email send failed")
		}
	}()
	return nil
}

func (e *EmailSenderImpl) SendOTPEmail(ctx context.Context, toEmail string, otp string) error {
	_ = ctx
	return e.SendEmail(&model.SendEmailParams{
		Template: model.EmailTemplateOTP,
		To:       []model.EmailAddress{{Email: toEmail}},
		Params: map[string]any{
			"otp": otp,
		},
	})
}
