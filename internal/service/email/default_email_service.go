package email

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/rs/zerolog"
)

type DefaultEmailService struct {
	logger zerolog.Logger
}

func NewDefaultEmailService() *DefaultEmailService {
	return &DefaultEmailService{
		logger: logger.GetGoamLogger(),
	}
}

func (m *DefaultEmailService) SendEmail(tenant, realm string, email *model.SendEmailParams) error {
	m.logger.Info().Str("tenant", tenant).Str("realm", realm).Str("email", fmt.Sprintf("%+v", email)).Msg("sending email")
	return nil
}
