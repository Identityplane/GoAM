package email

import (
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
)

func TestResolveMessageContent_templateOTP(t *testing.T) {
	subject, txt, html, err := resolveMessageContent("acme", "customers", &model.SendEmailParams{
		Template: model.EmailTemplateOTP,
		Params:   map[string]any{"otp": "123456"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if subject != "Your verification code" {
		t.Fatalf("subject %q", subject)
	}
	if !strings.Contains(txt, "123456") || !strings.Contains(html, "123456") {
		t.Fatalf("otp not in bodies: txt=%q html=%q", txt, html)
	}
}

func TestResolveMessageContent_rawRequiresSubject(t *testing.T) {
	_, _, _, err := resolveMessageContent("t", "r", &model.SendEmailParams{
		TextBody: "x",
	})
	if err == nil || !strings.Contains(err.Error(), "subject") {
		t.Fatalf("expected subject error, got %v", err)
	}
}

func TestResolveMessageContent_rawOK(t *testing.T) {
	subject, txt, html, err := resolveMessageContent("t", "r", &model.SendEmailParams{
		Subject:  "Subj",
		TextBody: "plain",
		HTMLBody: "<b>x</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if subject != "Subj" || txt != "plain" || html != "<b>x</b>" {
		t.Fatalf("got %q %q %q", subject, txt, html)
	}
}

func TestDefaultEmailService_SendEmail_noSMTP(t *testing.T) {
	svc := NewDefaultEmailService(SmtpConfig{})
	err := svc.SendEmail("t", "r", &model.SendEmailParams{
		Template: model.EmailTemplateOTP,
		To:       []model.EmailAddress{{Email: "u@x.test"}},
		Params:   map[string]any{"otp": "000000"},
	})
	if err == nil || !strings.Contains(err.Error(), "smtp host") {
		t.Fatalf("expected smtp host error, got %v", err)
	}
}

func TestGenericHTMLTextEmailParams(t *testing.T) {
	p := GenericHTMLTextEmailParams("Hi", []model.EmailAddress{{Email: "a@b.c", Name: "A"}},
		[]model.EmailAddress{{Email: "cc@b.c"}}, nil, "t", "<p>h</p>")
	if p.Subject != "Hi" || len(p.Cc) != 1 {
		t.Fatalf("%+v", p)
	}
}
