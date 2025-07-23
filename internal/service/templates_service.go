package service

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// Responseible for returning the right templates for a given realm
// this should be very fast as it is called for every request
type TemplatesService interface {
	GetTemplates(tenant, realm, flowId, nodeName string) (*template.Template, error)
	GetErrorTemplate(tenant, realm, flowId string) (*template.Template, error)
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
	baseTemplates *template.Template
	errorTemplate *template.Template

	// For each node we have a template to render that includes all the components and the base templates
	nodeTemplates map[string]*template.Template
)

func (s *templatesService) GetTemplates(tenant, realm, flowId, nodeName string) (*template.Template, error) {

	// In the future we might want to create overwrites for templates per tenant, realm or flowId
	// but for now we just ignore these params

	// Get the node template for the current node
	filename := fmt.Sprintf("%s.html", nodeName)
	nodeTemplate, ok := nodeTemplates[filename]
	if !ok {
		return nil, fmt.Errorf("node template not found for nodeName: %s", nodeName)
	}
	return nodeTemplate, nil
}

// InitTemplates loads and parses the base layout template and all components
func (s *templatesService) initTemplates() error {

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
	loadNodeTemplates()

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

			// Clone the base template
			tmpl, err := baseTemplates.Clone()
			if err != nil {
				return fmt.Errorf("failed to clone base template: %w", err)
			}

			// Parse the node template
			_, err = tmpl.ParseFS(templatesFS, nodeTemplatePath)
			if err != nil {
				return fmt.Errorf("failed to parse node template %s: %w", entry.Name(), err)
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
