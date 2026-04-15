package email

import (
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/rs/zerolog"
)

// DefaultEmailService sends mail over SMTP using built-in templates or raw bodies.
type DefaultEmailService struct {
	logger zerolog.Logger
	cfg    SmtpConfig
}

func NewDefaultEmailService(cfg SmtpConfig) *DefaultEmailService {
	return &DefaultEmailService{
		logger: logger.GetGoamLogger(),
		cfg:    cfg,
	}
}

func (m *DefaultEmailService) SendEmail(tenant, realm string, email *model.SendEmailParams) error {
	if email == nil {
		return fmt.Errorf("email params are nil")
	}
	if strings.TrimSpace(m.cfg.Host) == "" {
		return fmt.Errorf("smtp host is not configured")
	}
	if strings.TrimSpace(m.cfg.FromEmail) == "" && (email.From == nil || strings.TrimSpace(email.From.Email) == "") {
		return fmt.Errorf("smtp from address is not configured")
	}

	subject, textBody, htmlBody, err := resolveMessageContent(tenant, realm, email)
	if err != nil {
		return err
	}

	from := model.EmailAddress{Email: m.cfg.FromEmail, Name: m.cfg.FromName}
	if email.From != nil && strings.TrimSpace(email.From.Email) != "" {
		from = *email.From
	}

	msg, err := buildRFC5322Message(subject, from, email.ReplyTo, email.To, email.Cc, email.Bcc, textBody, htmlBody)
	if err != nil {
		return err
	}

	rcpts := allEnvelopeRecipients(email.To, email.Cc, email.Bcc)
	envFrom := envelopeFromEmail(from)
	port := m.cfg.Port
	if port <= 0 {
		port = 587
	}

	if err := deliverSMTP(m.cfg.Host, port, m.cfg.UseImplicitTLS, m.cfg.Username, m.cfg.Password, envFrom, rcpts, msg); err != nil {
		return err
	}

	m.logger.Info().
		Str("tenant", tenant).
		Str("realm", realm).
		Str("subject", subject).
		Strs("rcpt", rcpts).
		Msg("email sent")
	return nil
}

func resolveMessageContent(tenant, realm string, p *model.SendEmailParams) (subject, textBody, htmlBody string, err error) {
	switch {
	case strings.TrimSpace(p.Template) != "":
		subject = subjectForBuiltInTemplate(p.Template)
		if subject == "" {
			return "", "", "", fmt.Errorf("unknown email template %q", p.Template)
		}
		if strings.TrimSpace(p.Subject) != "" {
			subject = p.Subject
		}
		params := p.Params
		if params == nil {
			params = map[string]any{}
		}
		textBody, htmlBody, err = renderTemplateBodies(p.Template, tenant, realm, params)
		if err != nil {
			return "", "", "", err
		}
		return subject, textBody, htmlBody, nil

	case strings.TrimSpace(p.TextBody) != "" || strings.TrimSpace(p.HTMLBody) != "":
		if strings.TrimSpace(p.Subject) == "" {
			return "", "", "", fmt.Errorf("subject is required when sending raw html/text email")
		}
		return p.Subject, p.TextBody, p.HTMLBody, nil

	default:
		return "", "", "", fmt.Errorf("either template or text/html body must be set")
	}
}
