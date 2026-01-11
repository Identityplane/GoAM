package service

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTempalte(t *testing.T) {

	// Arrange
	service := NewTemplatesService().(*templatesService)
	view := &ViewData{
		Title:         "Test",
		NodeName:      "askEmail",
		Prompts:       map[string]string{},
		Debug:         false,
		Error:         "",
		State:         &model.AuthenticationSession{},
		StateJSON:     "",
		FlowName:      "flow1",
		Node:          &model.GraphNode{},
		StylePath:     "",
		ScriptPath:    "",
		Message:       "",
		CustomConfig:  map[string]string{},
		Tenant:        "acme",
		Realm:         "customers",
		FlowPath:      "flow1",
		LoginUri:      "",
		AssetsJSPath:  "",
		AssetsCSSPath: "",
		CspNonce:      "",
	}

	// Act
	tmpl, err := service.GetTemplates("acme", "customers", "flow1", "askEmail")
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	// Assert
	// Execute the template
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", view); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	// Assert
	assert.Contains(t, output, "<label for=\"email\">Email</label>")
}

func TestLoadTempalteOverride(t *testing.T) {

	// Arrange
	service := NewTemplatesService().(*templatesService)
	view := &ViewData{
		Title:         "Test",
		NodeName:      "askEmail",
		Prompts:       map[string]string{},
		Debug:         false,
		Error:         "",
		State:         &model.AuthenticationSession{},
		StateJSON:     "",
		FlowName:      "flow1",
		Node:          &model.GraphNode{},
		StylePath:     "",
		ScriptPath:    "",
		Message:       "",
		CustomConfig:  map[string]string{},
		Tenant:        "acme",
		Realm:         "customers",
		FlowPath:      "flow1",
		LoginUri:      "",
		AssetsJSPath:  "",
		AssetsCSSPath: "",
		CspNonce:      "",
	}

	// Act
	override := `{{ define "content" }}
OVERRIDE
{{ end }}`

	err := service.CreateTemplateOverride("acme", "customers", "flow1", "askEmail", override)
	if err != nil {
		t.Fatalf("failed to create template override: %v", err)
	}

	// Check if override was created
	overrides := service.ListTemplateOverrides()
	if len(overrides) == 0 {
		t.Fatalf("override was not created")
	}

	tmpl, err := service.GetTemplates("acme", "customers", "flow1", "askEmail")
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	// Assert
	// Execute the template
	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", view); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	// Assert
	assert.Contains(t, output, "OVERRIDE")
}

func TestLayoutTemplateOverride(t *testing.T) {

	// Arrange
	service := NewTemplatesService().(*templatesService)
	view := &ViewData{
		Title:         "Test",
		NodeName:      "askEmail",
		Prompts:       map[string]string{},
		Debug:         false,
		Error:         "",
		State:         &model.AuthenticationSession{},
		StateJSON:     "",
		FlowName:      "flow1",
		Node:          &model.GraphNode{},
		StylePath:     "",
		ScriptPath:    "",
		Message:       "",
		CustomConfig:  map[string]string{},
		Tenant:        "acme",
		Realm:         "customers",
		FlowPath:      "flow1",
		LoginUri:      "",
		AssetsJSPath:  "",
		AssetsCSSPath: "",
		CspNonce:      "",
	}

	// Act - Override the layout template
	layoutOverride := `{{ define "layout" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
</head>
<body>
  <div class="custom-layout">
    {{ template "content" . }}
  </div>
</body>
</html>
{{ end }}

{{ define "content" }}
CUSTOM LAYOUT OVERRIDE
{{ end }}`

	err := service.CreateTemplateOverride("acme", "customers", "flow1", "layout", layoutOverride)
	if err != nil {
		t.Fatalf("failed to create layout template override: %v", err)
	}

	tmpl, err := service.GetTemplates("acme", "customers", "flow1", "layout")
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	// Assert
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", view); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	// Assert
	assert.Contains(t, output, "CUSTOM LAYOUT OVERRIDE")
	assert.Contains(t, output, "custom-layout")
}

func TestErrorTemplateOverride(t *testing.T) {

	// Arrange
	service := NewTemplatesService().(*templatesService)
	view := &ViewData{
		Title:         "Error Test",
		NodeName:      "error",
		Prompts:       map[string]string{},
		Debug:         false,
		Error:         "Test error message",
		State:         &model.AuthenticationSession{},
		StateJSON:     "",
		FlowName:      "flow1",
		Node:          &model.GraphNode{},
		StylePath:     "",
		ScriptPath:    "",
		Message:       "",
		CustomConfig:  map[string]string{},
		Tenant:        "acme",
		Realm:         "customers",
		FlowPath:      "flow1",
		LoginUri:      "",
		AssetsJSPath:  "",
		AssetsCSSPath: "",
		CspNonce:      "",
	}

	// Act - Override the error template
	errorOverride := `{{ define "error" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
</head>
<body>
  <div class="custom-error">
    <h1>CUSTOM ERROR PAGE</h1>
    <p>{{ .Error }}</p>
  </div>
</body>
</html>
{{ end }}`

	err := service.CreateTemplateOverride("acme", "customers", "flow1", "error", errorOverride)
	if err != nil {
		t.Fatalf("failed to create error template override: %v", err)
	}

	tmpl, err := service.GetErrorTemplate("acme", "customers", "flow1")
	if err != nil {
		t.Fatalf("failed to load error template: %v", err)
	}

	// Assert
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "error", view); err != nil {
		t.Fatalf("failed to execute error template: %v", err)
	}

	output := buf.String()

	// Assert
	assert.Contains(t, output, "CUSTOM ERROR PAGE")
	assert.Contains(t, output, "Test error message")
	assert.Contains(t, output, "custom-error")
}

func TestRemoveTemplateOverride(t *testing.T) {

	// Arrange
	service := NewTemplatesService().(*templatesService)

	// Act - Create an override
	override := `{{ define "content" }}
OVERRIDE
{{ end }}`

	err := service.CreateTemplateOverride("acme", "customers", "flow1", "askEmail", override)
	if err != nil {
		t.Fatalf("failed to create template override: %v", err)
	}

	// Check that override exists
	overrides := service.ListTemplateOverrides()
	if len(overrides) == 0 {
		t.Fatalf("override was not created")
	}

	// Remove the override
	err = service.RemoveTemplateOverride("acme", "customers", "flow1", "askEmail")
	if err != nil {
		t.Fatalf("failed to remove template override: %v", err)
	}

	// Check that override was removed
	overrides = service.ListTemplateOverrides()
	if len(overrides) != 0 {
		t.Fatalf("override was not removed")
	}

	// Try to remove non-existent override
	err = service.RemoveTemplateOverride("acme", "customers", "flow1", "nonexistent")
	if err == nil {
		t.Fatalf("expected error when removing non-existent override")
	}
}

func TestMultipleTemplateOverrides(t *testing.T) {

	// Arrange
	service := NewTemplatesService().(*templatesService)
	view := &ViewData{
		Title:         "Test",
		NodeName:      "askEmail",
		Prompts:       map[string]string{},
		Debug:         false,
		Error:         "",
		State:         &model.AuthenticationSession{},
		StateJSON:     "",
		FlowName:      "flow1",
		Node:          &model.GraphNode{},
		StylePath:     "",
		ScriptPath:    "",
		Message:       "",
		CustomConfig:  map[string]string{},
		Tenant:        "acme",
		Realm:         "customers",
		FlowPath:      "flow1",
		LoginUri:      "",
		AssetsJSPath:  "",
		AssetsCSSPath: "",
		CspNonce:      "",
	}

	// Act - Create multiple overrides
	override1 := `{{ define "content" }}
OVERRIDE 1
{{ end }}`

	override2 := `{{ define "content" }}
OVERRIDE 2
{{ end }}`

	err := service.CreateTemplateOverride("acme", "customers", "flow1", "askEmail", override1)
	if err != nil {
		t.Fatalf("failed to create first template override: %v", err)
	}

	err = service.CreateTemplateOverride("acme", "customers", "flow2", "askEmail", override2)
	if err != nil {
		t.Fatalf("failed to create second template override: %v", err)
	}

	// Check that both overrides exist
	overrides := service.ListTemplateOverrides()
	if len(overrides) != 2 {
		t.Fatalf("expected 2 overrides, got %d", len(overrides))
	}

	// Test first override
	tmpl1, err := service.GetTemplates("acme", "customers", "flow1", "askEmail")
	if err != nil {
		t.Fatalf("failed to load first template: %v", err)
	}

	var buf1 bytes.Buffer
	if err := tmpl1.ExecuteTemplate(&buf1, "layout", view); err != nil {
		t.Fatalf("failed to execute first template: %v", err)
	}

	output1 := buf1.String()
	assert.Contains(t, output1, "OVERRIDE 1")

	// Test second override
	tmpl2, err := service.GetTemplates("acme", "customers", "flow2", "askEmail")
	if err != nil {
		t.Fatalf("failed to load second template: %v", err)
	}

	var buf2 bytes.Buffer
	if err := tmpl2.ExecuteTemplate(&buf2, "layout", view); err != nil {
		t.Fatalf("failed to execute second template: %v", err)
	}

	output2 := buf2.String()
	assert.Contains(t, output2, "OVERRIDE 2")
}

func TestQrCodeFunction(t *testing.T) {
	// Test the qrCodeToBase64 function directly
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, result template.URL)
	}{
		{
			name:  "valid URL generates QR code",
			input: "http://localhost:8081/acme/customers/auth/qrWebToMobile/action/0ec140d9-31a8-4c14-bbd0-fca765e7b356",
			validate: func(t *testing.T, result template.URL) {
				resultStr := string(result)
				assert.NotEmpty(t, resultStr)
				assert.True(t, strings.HasPrefix(resultStr, "data:image/png;base64,"), "should start with data URL prefix")
				// Verify it's valid base64
				base64Part := strings.TrimPrefix(resultStr, "data:image/png;base64,")
				decoded, err := base64.StdEncoding.DecodeString(base64Part)
				assert.NoError(t, err)
				assert.NotEmpty(t, decoded)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qrCodeToBase64(tt.input)

			tt.validate(t, result)
		})
	}
}

func TestQrCodeTemplateFunction(t *testing.T) {
	// Test the qrCode function in a template context
	service := NewTemplatesService().(*templatesService)

	// Create a test template that uses the qrCode function
	testTemplate := `{{ define "content" }}
<img src="{{ qrCode .Prompts.qr_code }}" alt="QR Code" />
{{ end }}`

	// Create an override with the test template
	err := service.CreateTemplateOverride("test", "test", "test", "qrWebToMobile", testTemplate)
	require.NoError(t, err)

	view := &ViewData{
		Title:         "QR Test",
		NodeName:      "qrWebToMobile",
		Prompts:       map[string]string{"qr_code": "contentofqrcode"},
		Debug:         false,
		Error:         "",
		State:         &model.AuthenticationSession{},
		StateJSON:     "",
		FlowName:      "test",
		Node:          &model.GraphNode{},
		StylePath:     "",
		ScriptPath:    "",
		Message:       "",
		CustomConfig:  map[string]string{},
		Tenant:        "test",
		Realm:         "test",
		FlowPath:      "test",
		LoginUri:      "",
		AssetsJSPath:  "",
		AssetsCSSPath: "",
		CspNonce:      "",
	}

	// Get the template
	tmpl, err := service.GetTemplates("test", "test", "test", "qrWebToMobile")
	require.NoError(t, err)

	// Execute the template
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", view)
	require.NoError(t, err)

	output := buf.String()

	// Verify the QR code was generated
	assert.Contains(t, output, "data:image/png;base64,")
	assert.NotContains(t, output, "#ZgotmplZ", "should not contain template error marker")
	assert.NotContains(t, output, "http://localhost:8081", "should not contain the raw URL in img src")
}
