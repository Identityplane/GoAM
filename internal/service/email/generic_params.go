package email

import "github.com/Identityplane/GoAM/pkg/model"

// GenericHTMLTextEmailParams returns a SendEmailParams for a multipart (or single-part)
// message without using a named template.
func GenericHTMLTextEmailParams(subject string, to, cc, bcc []model.EmailAddress, textBody, htmlBody string) *model.SendEmailParams {
	return &model.SendEmailParams{
		Subject:  subject,
		To:       to,
		Cc:       cc,
		Bcc:      bcc,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}
}
