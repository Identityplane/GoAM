package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"sync"
	texttemplate "text/template"

	"github.com/Identityplane/GoAM/pkg/model"
)

//go:embed templates/*
var builtinTemplates embed.FS

var (
	tplMu     sync.Mutex
	htmlCache = map[string]*template.Template{}
	textCache = map[string]*texttemplate.Template{}
)

func renderTemplateBodies(templateKey, tenant, realm string, params map[string]any) (textBody, htmlBody string, err error) {
	data := map[string]any{
		"tenant": tenant,
		"realm":  realm,
	}
	for k, v := range params {
		data[k] = v
	}

	htmlT, err := parseHTMLTemplate(templateKey)
	if err != nil {
		return "", "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlT.Execute(&htmlBuf, data); err != nil {
		return "", "", fmt.Errorf("execute html template %q: %w", templateKey, err)
	}

	txtT, err := parseTextTemplate(templateKey)
	if err != nil {
		return "", "", err
	}
	var txtBuf bytes.Buffer
	if err := txtT.Execute(&txtBuf, data); err != nil {
		return "", "", fmt.Errorf("execute text template %q: %w", templateKey, err)
	}

	return txtBuf.String(), htmlBuf.String(), nil
}

func parseHTMLTemplate(name string) (*template.Template, error) {
	tplMu.Lock()
	defer tplMu.Unlock()
	if t, ok := htmlCache[name]; ok {
		return t, nil
	}
	path := fmt.Sprintf("templates/%s.html", name)
	raw, err := builtinTemplates.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	t, err := template.New(path).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse html %s: %w", name, err)
	}
	htmlCache[name] = t
	return t, nil
}

func parseTextTemplate(name string) (*texttemplate.Template, error) {
	tplMu.Lock()
	defer tplMu.Unlock()
	if t, ok := textCache[name]; ok {
		return t, nil
	}
	path := fmt.Sprintf("templates/%s.txt", name)
	raw, err := builtinTemplates.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	t, err := texttemplate.New(path).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse text %s: %w", name, err)
	}
	textCache[name] = t
	return t, nil
}

func subjectForBuiltInTemplate(name string) string {
	switch name {
	case model.EmailTemplateOTP:
		return "Your verification code"
	case model.EmailTemplateWelcome:
		return "Welcome"
	default:
		return ""
	}
}
