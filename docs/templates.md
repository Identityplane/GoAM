# Template System Documentation

## Overview

The template system in GoAM provides a flexible way to render HTML pages for authentication flows. It supports both static templates and dynamic overrides that can be customized per tenant, realm, and flow.

## Template Structure

### Base Templates
- **Layout Template** (`templates/layout.html`): The main layout template that wraps all content
- **Error Template** (`templates/error.html`): Template for error pages
- **Node Templates** (`templates/nodes/`): Individual page templates for different authentication steps

### Components
- **Components** (`templates/components/`): Reusable template components like debug panels, forms, etc.

## Template Overrides

### Dynamic Overrides

You can create template overrides programmatically using the `TemplatesService`:

```go
service := NewTemplatesService()

// Override a specific node template
err := service.CreateTemplateOverride("tenant", "realm", "flowId", "nodeName", templateString)

// Remove an override
err := service.RemoveTemplateOverride("tenant", "realm", "flowId", "nodeName")

// List all overrides
overrides := service.ListTemplateOverrides()
```

### Static File Overrides

You can also load template overrides from static files:

#### Local Filesystem
```go
// Load overrides from a local directory
err := service.LoadTemplateOverridesFromPath("tenant", "realm", "/path/to/templates")
```

#### Embedded Filesystem
```go
// Load overrides from an embedded filesystem (e.g., embed.FS)
err := service.LoadTemplateOverridesFromFS("tenant", "realm", embeddedFS, "templates")
```

### Override Precedence

Template overrides follow this precedence order:
1. Dynamic overrides (created via `CreateTemplateOverride`)
2. Static file overrides (loaded via `LoadTemplateOverridesFromPath` or `LoadTemplateOverridesFromFS`)
3. Default embedded templates

### Override Key Format

Overrides are stored using the key format: `tenant/realm/flowId/templateName`

- **Content templates**: `tenant/realm/flowId/nodeName` (e.g., `acme/customers/flow1/askEmail`)
- **Layout templates**: `tenant/realm/flowId/layout` (e.g., `acme/customers/flow1/layout`)
- **Error templates**: `tenant/realm/flowId/error` (e.g., `acme/customers/flow1/error`)

## ViewData

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

## Template Functions

### Available Functions
- `title(string)`: Capitalizes the first letter of a string

## Best Practices

### Creating Overrides

1. **Content Templates**: Define a `content` template block
   ```html
   {{ define "content" }}
   <div class="custom-content">
     <h1>{{ .Title }}</h1>
     <!-- Your custom content here -->
   </div>
   {{ end }}
   ```

2. **Layout Templates**: Define a complete `layout` template
   ```html
   {{ define "layout" }}
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
   ```

3. **Error Templates**: Define a complete error page
   ```html
   {{ define "error" }}
   <!DOCTYPE html>
   <html>
   <head>
     <title>Error - {{ .Title }}</title>
   </head>
   <body>
     <div class="error-page">
       <h1>Error</h1>
       <p>{{ .Error }}</p>
     </div>
   </body>
   </html>
   {{ end }}
   ```

### File Organization

When using static file overrides, organize your templates like this:
```
config/tenants/acme/customers/templates/
├── layout.html          # Custom layout for acme/customers
├── askEmail.html        # Custom email form for acme/customers
├── askPassword.html     # Custom password form for acme/customers
└── error.html          # Custom error page for acme/customers
```

### Testing Overrides

Always test your template overrides to ensure they render correctly:

```go
// Test that an override is applied
tmpl, err := service.GetTemplates("tenant", "realm", "flowId", "nodeName")
if err != nil {
    // Handle error
}

var buf bytes.Buffer
err = tmpl.ExecuteTemplate(&buf, "layout", viewData)
if err != nil {
    // Handle error
}

output := buf.String()
// Assert that your custom content is present
```

## Integration with Embed.FS

The template system supports both local filesystem and embedded filesystem overrides. This is particularly useful for:

1. **Testing**: Use embedded filesystems in tests for consistent behavior
2. **Deployment**: Embed custom templates into the binary for deployment
3. **Flexibility**: Choose between file-based and embedded overrides based on your needs

Example with embed.FS:
```go
//go:embed custom_templates/*
var customTemplates embed.FS

// Load overrides from embedded filesystem
err := service.LoadTemplateOverridesFromFS("tenant", "realm", customTemplates, "custom_templates")
```
