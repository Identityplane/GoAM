package services

import "github.com/Identityplane/GoAM/pkg/model"

type EmailService interface {
	SendEmail(tenant, realm string, email *model.SendEmailParams) error
}
