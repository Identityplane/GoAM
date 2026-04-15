package model

import "context"

const (
	EmailTemplateOTP     = "email-otp"
	EmailTemplateWelcome = "welcome"
)

// SendEmailParams describes an outgoing email. Use either Template + Params for
// built-in templates, or Subject + TextBody/HTMLBody for ad-hoc messages.
type SendEmailParams struct {
	Template string
	Subject  string

	To  []EmailAddress
	Cc  []EmailAddress
	Bcc []EmailAddress

	ReplyTo *EmailAddress
	From    *EmailAddress

	TextBody string
	HTMLBody string

	Params map[string]any
}

type EmailAddress struct {
	Email string
	Name  string
}

// EmailSender delivers transactional email for a realm.
type EmailSender interface {
	SendEmail(email *SendEmailParams) error
	// SendOTPEmail is a narrow entry point for one-time codes (easier to mock in OTP tests).
	SendOTPEmail(ctx context.Context, toEmail string, otp string) error
}
