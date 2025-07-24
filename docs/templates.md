# Template System Documentation

## Overview

The GoAM template system provides a flexible way to render authentication flows with support for custom overrides. Templates are organized in a hierarchical structure that allows for both standard templates and custom overrides per tenant, realm, and flow.

## Template Structure

### Directory Layout

```
internal/service/templates/
├── layout.html          # Base layout template
├── error.html           # Error page template
├── nodes/               # Node-specific templates
│   ├── askEmail.html
│   ├── askPassword.html
│   ├── askUsername.html
│   ├── emailOTP.html
│   ├── hcaptcha.html
│   ├── messageConfirmation.html
│   ├── onboardingWithPasskey.html
│   ├── passwordOrSocialLogin.html
│   ├── registerPasskey.html
│   ├── successResult.html
│   ├── failureResult.html
│   ├── telegramLogin.html
│   └── verifyPasskey.html
└── components/          # Reusable components
    ├── debug.html
    ├── loginWithGoogle.html
    └── loginWithGitHub.html
```

### Template Hierarchy

1. **Layout Template** (`layout.html`): The main wrapper that provides the HTML structure
2. **Node Templates** (`nodes/*.html`): Content templates for specific authentication nodes
3. **Components** (`components/*.html`): Reusable template components
4. **Error Template** (`error.html`): Error page template

### Template Structure

#### Layout Template
The layout template provides the basic HTML structure and includes:
- HTML head with title, CSS, and JavaScript
- Main content area where node templates are rendered
- Debug section (when debug mode is enabled)

```html
{{ define "layout" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="{{ .AssetsCSSPath }}">
  <script defer src="{{ .AssetsJSPath }}" nonce="{{ .CspNonce }}"></script>
  <link rel="stylesheet" href="{{ .StylePath }}">
</head>
<body>
  <div class="main-content" data-node="{{ .Node.Use }}">
    <div class="login-container">
      {{ if .Error }}
      <div class="error" style="color: red; font-weight: bold;">
        {{ .Error }}
      </div>
      {{ end }}
      {{ template "content" . }}
    </div>
  </div>
  {{ template "debug" . }}
</body>
</html>
{{ end }}
```

#### Node Templates
Node templates define the content for specific authentication steps:

```html
{{ define "content" }}
<form method="POST" class="login-form" action="{{ .LoginUri}}">
  <h2>Welcome to GoAM</h2>
  <p>{{index .CustomConfig "message"}}</p>

  <div class="input-group">
    <input type="hidden" name="step" value="{{ .NodeName }}">
    <label for="email">Email</label>
    <input type="email" name="email" id="email" placeholder="email" required />
  </div>
  <button type="submit">Login</button>
</form>
{{ end }}
```

#### Components
Components are reusable template parts that can be included in other templates:

```html
{{ define "debug" }}
{{ if .Debug }}
<div class="debug">
  <hr><h3>Debug State</h3>
  <div class="json-viewer">
    <pre id="debug-json">{{ .StateJSON }}</pre>
  </div>
</div>
{{ end }}
{{ end }}
```

## Template Override System

The template system supports dynamic overrides that can be applied per tenant, realm, and flow. This allows for customization without modifying the base templates.

### Override Key Structure

Overrides are identified by a key in the format: `{tenant}/{realm}/{flowId}/{nodeName}`

Examples:
- `acme/customers/flow1/askEmail` - Override for email input in ACME tenant, customers realm, flow1
- `acme/customers/flow1/layout` - Override for the entire layout
- `acme/customers/flow1/error` - Override for error pages

### Creating Template Overrides

#### Content Override
To override a specific node's content:

```go
override := `{{ define "content" }}
<div class="custom-login">
  <h1>Custom Login Form</h1>
  <form method="POST" action="{{ .LoginUri }}">
    <input type="hidden" name="step" value="{{ .NodeName }}">
    <label for="email">Email Address</label>
    <input type="email" name="email" required />
    <button type="submit">Sign In</button>
  </form>
</div>
{{ end }}`

err := service.CreateTemplateOverride("acme", "customers", "flow1", "askEmail", override)
```

#### Layout Override
To override the entire layout:

```go
layoutOverride := `{{ define "layout" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="/custom/styles.css">
</head>
<body>
  <div class="custom-layout">
    {{ template "content" . }}
  </div>
</body>
</html>
{{ end }}

{{ define "content" }}
CUSTOM LAYOUT CONTENT
{{ end }}`

err := service.CreateTemplateOverride("acme", "customers", "flow1", "layout", layoutOverride)
```

#### Error Template Override
To override error pages:

```go
errorOverride := `{{ define "error" }}
<!DOCTYPE html>
<html>
<head>
  <title>{{ .Title }}</title>
</head>
<body>
  <div class="custom-error">
    <h1>Custom Error Page</h1>
    <p>{{ .Error }}</p>
  </div>
</body>
</html>
{{ end }}`

err := service.CreateTemplateOverride("acme", "customers", "flow1", "error", errorOverride)
```

### Managing Template Overrides

#### List All Overrides
```go
overrides := service.ListTemplateOverrides()
for key := range overrides {
    fmt.Printf("Override: %s\n", key)
}
```

#### Remove an Override
```go
err := service.RemoveTemplateOverride("acme", "customers", "flow1", "askEmail")
if err != nil {
    // Handle error
}
```

### Override Precedence

1. **Custom Override**: If a custom override exists for the specific tenant/realm/flow/node combination, it will be used
2. **Standard Template**: If no override exists, the standard template from the filesystem will be used

### Template Data

All templates receive a `ViewData` struct with the following fields:

```go
type ViewData struct {
    Title         string
    NodeName      string
    Prompts       map[string]string
    Debug         bool
    Error         string
    State         *model.AuthenticationSession
    StateJSON     string
    FlowName      string
    Node          *model.GraphNode
    StylePath     string
    ScriptPath    string
    Message       string
    CustomConfig  map[string]string
    Tenant        string
    Realm         string
    FlowPath      string
    LoginUri      string
    AssetsJSPath  string
    AssetsCSSPath string
    CspNonce      string
}
```

### Template Functions

Templates have access to the following functions:

- `title(s string)`: Capitalizes the first letter of a string

### Best Practices

1. **Use Semantic HTML**: Ensure your templates use proper HTML structure
2. **Include Required Fields**: Always include hidden fields like `step` with the node name
3. **Handle Errors**: Use the `.Error` field to display error messages
4. **Support Debug Mode**: Include debug information when `.Debug` is true
5. **Use Components**: Leverage the component system for reusable parts
6. **Test Overrides**: Always test your overrides to ensure they work correctly

### Example: Complete Custom Login Form

```go
customLogin := `{{ define "content" }}
<div class="custom-login-container">
  <div class="login-header">
    <h1>Welcome Back</h1>
    <p>Please sign in to continue</p>
  </div>
  
  {{ if .Error }}
  <div class="error-message">
    {{ .Error }}
  </div>
  {{ end }}
  
  <form method="POST" class="login-form" action="{{ .LoginUri }}">
    <input type="hidden" name="step" value="{{ .NodeName }}">
    
    <div class="form-group">
      <label for="email">Email Address</label>
      <input type="email" name="email" id="email" 
             placeholder="Enter your email" required />
    </div>
    
    <button type="submit" class="btn-primary">
      Sign In
    </button>
  </form>
  
  <div class="login-footer">
    <p>Don't have an account? <a href="/register">Sign up</a></p>
  </div>
</div>
{{ end }}`

err := service.CreateTemplateOverride("acme", "customers", "flow1", "askEmail", customLogin)
```

This template system provides a powerful and flexible way to customize authentication flows while maintaining the core functionality and security of the GoAM system.
