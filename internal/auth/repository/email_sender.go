package repository

import (
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

	return e.emailService.SendEmail(e.tenant, e.realm, email)
}
