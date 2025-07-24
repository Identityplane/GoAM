package service

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/Identityplane/GoAM/pkg/model"
)

// ViewData is passed to all templates for dynamic rendering
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

// Responseible for returning the right templates for a given realm
// this should be very fast as it is called for every request
type TemplatesService interface {
	GetTemplates(tenant, realm, flowId, nodeName string) (*template.Template, error)
	GetErrorTemplate(tenant, realm, flowId string) (*template.Template, error)
	CreateTemplateOverride(tenant, realm, flowId, nodeName, templateString string) error
	RemoveTemplateOverride(tenant, realm, flowId, nodeName string) error
	ListTemplateOverrides() map[string]bool
}

type templatesService struct {
}

func NewTemplatesService() TemplatesService {

	service := &templatesService{}

	if err := service.initTemplates(); err != nil {
		panic("failed to initialize templates: " + err.Error())
	}

	return service
}

//go:embed templates/*
var templatesFS embed.FS

// Configuration
var (
	LayoutTemplatePath = "templates/layout.html"
	ErrorTemplatePath  = "templates/error.html"
	NodeTemplatesPath  = "templates/nodes"
	ComponentsPath     = "templates/components"
)

// loaded tempaltes
var (
	// The standard templates
	baseTemplates *template.Template
	errorTemplate *template.Template
	nodeTemplates map[string]*template.Template

	// The templates that are overwritten - store the template strings
	overwriteTemplates map[string]string
)

// CreateTemplateOverride creates a template override for a given node
func (s *templatesService) CreateTemplateOverride(tenant, realm, flowId, nodeName, templateString string) error {

	// Validate input parameters
	if tenant == "" || realm == "" || flowId == "" || nodeName == "" {
		return fmt.Errorf("tenant, realm, flowId, and nodeName cannot be empty")
	}

	if templateString == "" {
		return fmt.Errorf("templateString cannot be empty")
	}

	overwriteIndex := fmt.Sprintf("%s/%s/%s/%s", tenant, realm, flowId, nodeName)

	// Add the template to the map
	overwriteTemplates[overwriteIndex] = templateString

	return nil
}

// RemoveTemplateOverride removes a template override for a given node
func (s *templatesService) RemoveTemplateOverride(tenant, realm, flowId, nodeName string) error {

	overwriteIndex := fmt.Sprintf("%s/%s/%s/%s", tenant, realm, flowId, nodeName)

	// Check if the override exists
	if _, exists := overwriteTemplates[overwriteIndex]; !exists {
		return fmt.Errorf("template override not found for: %s", overwriteIndex)
	}

	// Remove the template from the map
	delete(overwriteTemplates, overwriteIndex)

	return nil
}

// ListTemplateOverrides returns a map of all existing template overrides
func (s *templatesService) ListTemplateOverrides() map[string]bool {
	result := make(map[string]bool)
	for key := range overwriteTemplates {
		result[key] = true
	}
	return result
}

// GetTemplates returns the template for a given node
func (s *templatesService) GetTemplates(tenant, realm, flowId, nodeName string) (*template.Template, error) {

	// Check for template override first
	overwriteIndex := fmt.Sprintf("%s/%s/%s/%s", tenant, realm, flowId, nodeName)
	if overrideTemplateString, exists := overwriteTemplates[overwriteIndex]; exists {
		// Create a new template with the same functions as base templates
		result := template.New("override").Funcs(template.FuncMap{
			"title": title,
		})

		// Parse the layout template
		_, err := result.ParseFS(templatesFS, LayoutTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse layout template: %w", err)
		}

		// Load components
		if err := loadComponents(result); err != nil {
			return nil, fmt.Errorf("failed to load components: %w", err)
		}

		// Parse the override template string directly into the result template
		_, err = result.Parse(overrideTemplateString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse override template: %w", err)
		}

		return result, nil
	}

	// Get the node template for the current node
	filename := fmt.Sprintf("%s.html", nodeName)
	nodeTemplate, ok := nodeTemplates[filename]

	if !ok {
		return nil, fmt.Errorf("node template not found for nodeName: %s", nodeName)
	}

	// Create a fresh template by cloning the node template and adding the layout
	result := template.Must(nodeTemplate.Clone())

	// Add the base templates to the node template
	_, err := result.AddParseTree("layout", baseTemplates.Tree)
	if err != nil {
		return nil, fmt.Errorf("failed to add parse tree: %w", err)
	}

	return result, nil
}

// InitTemplates loads and parses the base layout template and all components
func (s *templatesService) initTemplates() error {

	// Initialize the overwrite templates map
	overwriteTemplates = make(map[string]string)

	// STEP 1: Parse the base templates. Those are the templates that are used for every page
	// Parse the base layout template
	var err error
	baseTemplates, err = template.New("layout").Funcs(template.FuncMap{
		"title": title,
	}).ParseFS(templatesFS, LayoutTemplatePath)

	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	// Load all component templates
	if err := loadComponents(baseTemplates); err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// STEP 2: Parse all node templates and put them in a map with their name as the key
	if err := loadNodeTemplates(); err != nil {
		return fmt.Errorf("failed to load node templates: %w", err)
	}

	// STEP 3: Parse the error template
	// Parse the base layout template
	errorTemplate, err = template.New("error").Funcs(template.FuncMap{
		"title": title,
	}).ParseFS(templatesFS, ErrorTemplatePath)

	if err != nil {
		return fmt.Errorf("failed to parse base template: %w", err)
	}

	// Load all component templates
	if err := loadComponents(errorTemplate); err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	return nil

}

// Return the error template
func (s *templatesService) GetErrorTemplate(tenant, realm, flowId string) (*template.Template, error) {

	// Check for template override first
	overwriteIndex := fmt.Sprintf("%s/%s/%s/error", tenant, realm, flowId)
	if overrideTemplateString, exists := overwriteTemplates[overwriteIndex]; exists {
		// Create a new template with the same functions as base templates
		result := template.New("error").Funcs(template.FuncMap{
			"title": title,
		})

		// Parse the override template string directly into the result template
		_, err := result.Parse(overrideTemplateString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse error override template: %w", err)
		}

		return result, nil
	}

	return errorTemplate, nil
}

func loadNodeTemplates() error {

	nodeTemplates = make(map[string]*template.Template)

	entries, err := templatesFS.ReadDir(NodeTemplatesPath)
	if err != nil {
		return fmt.Errorf("failed to read node templates directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			nodeTemplatePath := filepath.Join(NodeTemplatesPath, entry.Name())

			// Create a new template with the same functions as base templates
			tmpl := template.New("content").Funcs(template.FuncMap{
				"title": title,
			})

			// Parse the node template
			_, err := tmpl.ParseFS(templatesFS, nodeTemplatePath)
			if err != nil {
				return fmt.Errorf("failed to parse node template %s: %w", entry.Name(), err)
			}

			// Load components into the node template
			if err := loadComponents(tmpl); err != nil {
				return fmt.Errorf("failed to load components for node template %s: %w", entry.Name(), err)
			}

			// Add the node template to the map
			nodeTemplates[entry.Name()] = tmpl
		}
	}
	return nil
}

// loadComponents loads all component templates from the components directory
func loadComponents(tmpl *template.Template) error {

	entries, err := templatesFS.ReadDir(ComponentsPath)
	if err != nil {
		return fmt.Errorf("failed to read components directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {

			componentPath := filepath.Join(ComponentsPath, entry.Name())

			// Parse the component template directly into the main template
			_, err := tmpl.ParseFS(templatesFS, componentPath)
			if err != nil {
				return fmt.Errorf("failed to parse component %s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

func title(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0]-32) + s[1:] // capitalize first letter
}
