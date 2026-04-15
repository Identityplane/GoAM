package email

import "github.com/Identityplane/GoAM/pkg/server_settings"

// SmtpConfig holds outbound SMTP settings for DefaultEmailService.
type SmtpConfig struct {
	Host             string
	Port             int
	Username         string
	Password         string
	FromEmail        string
	FromName         string
	UseImplicitTLS   bool
}

// SmtpConfigFromServerSettings maps server configuration into SmtpConfig.
func SmtpConfigFromServerSettings(s *server_settings.GoamServerSettings) SmtpConfig {
	if s == nil {
		return SmtpConfig{}
	}
	return SmtpConfig{
		Host:             s.SmtpHost,
		Port:             s.SmtpPort,
		Username:         s.SmtpUsername,
		Password:         s.SmtpPassword,
		FromEmail:        s.SmtpFromEmail,
		FromName:         s.SmtpFromName,
		UseImplicitTLS:   s.SmtpUseImplicitTLS,
	}
}
