package service

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
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
	StaticPath    string
	AssetsCSSPath string
	CspNonce      string
}

type templatesService struct {
}

func NewTemplatesService() services_interface.TemplatesService {

	service := &templatesService{}

	// Initialize the maps
	componentTemplates = make(map[string]string)
	nodeTemplates = make(map[string]string)
	overwriteTemplates = make(map[string]string)

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
	layoutTemplate     string
	componentTemplates map[string]string
	nodeTemplates      map[string]string

	// The templates that are overwritten - store the template strings
	overwriteTemplates map[string]string

	log = logger.GetGoamLogger()
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

	log.Debug().Str("tenant", tenant).Str("realm", realm).Str("flowId", flowId).Str("nodeName", nodeName).Msg("created template override")

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

// LoadTemplateOverridesFromPath loads template overrides from a local file path for a specific tenant and realm
func (s *templatesService) LoadTemplateOverridesFromPath(tenant, realm, templatesPath string) error {

	// Validate input parameters
	if tenant == "" || realm == "" || templatesPath == "" {
		return fmt.Errorf("tenant, realm, and templatesPath cannot be empty")
	}

	// Check if the templates directory exists
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return fmt.Errorf("templates directory does not exist: %s", templatesPath)
	}

	// Use the local filesystem
	return s.LoadTemplateOverridesFromFS(tenant, realm, os.DirFS(templatesPath), ".")
}

// LoadTemplateOverridesFromFS loads template overrides from a filesystem for a specific tenant and realm
func (s *templatesService) LoadTemplateOverridesFromFS(tenant, realm string, templatesFS fs.FS, templatesPath string) error {

	// Validate input parameters
	if tenant == "" || realm == "" {
		return fmt.Errorf("tenant and realm cannot be empty")
	}

	// Read the directory contents
	entries, err := fs.ReadDir(templatesFS, templatesPath)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	// Process each entry
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Only process HTML files
		if !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		// Read the template file
		filePath := filepath.Join(templatesPath, entry.Name())
		templateContent, err := fs.ReadFile(templatesFS, filePath)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", filePath, err)
		}

		// Extract the template name from the filename (without .html extension)
		templateName := strings.TrimSuffix(entry.Name(), ".html")
		s.CreateTemplateOverride(tenant, realm, "*", templateName, string(templateContent))
	}

	return nil
}

// GetTemplates returns the template for a given node
func (s *templatesService) GetTemplates(tenant, realm, flowId, nodeName string) (*template.Template, error) {

	// Get the layout template
	usedLayout := s.findOverrideTemplate(tenant, realm, flowId, "layout")
	if usedLayout == "" {
		usedLayout = layoutTemplate
	}

	// Parse the layout template
	template, err := template.New("layout").Parse(usedLayout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout template: %w", err)
	}

	// Parse all components
	for _, componentTemplate := range componentTemplates {
		_, err := template.Parse(componentTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse component template: %w", err)
		}
	}

	// Parse the node template
	usedNode := s.findOverrideTemplate(tenant, realm, flowId, nodeName)
	if usedNode == "" {
		usedNode = nodeTemplates[nodeName+".html"]
	}

	// Parse the node template
	template, err = template.Parse(usedNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node template: %w", err)
	}

	return template, nil
}

func (s *templatesService) findOverrideTemplate(tenant, realm, flowId, nodeName string) string {

	// First check the most specific override
	overwriteIndex := fmt.Sprintf("%s/%s/%s/%s", tenant, realm, flowId, nodeName)
	if overrideTemplateString, exists := overwriteTemplates[overwriteIndex]; exists {
		return overrideTemplateString
	}

	// Check overwrite for all flows in the realm
	overwriteIndex = fmt.Sprintf("%s/%s/*/%s", tenant, realm, nodeName)
	if overrideTemplateString, exists := overwriteTemplates[overwriteIndex]; exists {
		return overrideTemplateString
	}

	// Overwride for all realms
	overwriteIndex = fmt.Sprintf("%s/*/*/%s", tenant, nodeName)
	if overrideTemplateString, exists := overwriteTemplates[overwriteIndex]; exists {
		return overrideTemplateString
	}

	// Overwride for all tenants
	overwriteIndex = fmt.Sprintf("*/*/*/%s", nodeName)
	if overrideTemplateString, exists := overwriteTemplates[overwriteIndex]; exists {
		return overrideTemplateString
	}

	return ""
}

// InitTemplates loads all the templates as strings
func (s *templatesService) initTemplates() error {

	// Load the layout template
	layoutTemplateString, err := templatesFS.ReadFile(LayoutTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read layout template: %w", err)
	}
	layoutTemplate = string(layoutTemplateString)

	// Read each component template
	entries, err := templatesFS.ReadDir(ComponentsPath)
	if err != nil {
		return fmt.Errorf("failed to read components directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			componentTemplatePath := filepath.Join(ComponentsPath, entry.Name())
			componentTemplateString, err := templatesFS.ReadFile(componentTemplatePath)
			if err != nil {
				return fmt.Errorf("failed to read component template %s: %w", entry.Name(), err)
			}
			componentTemplates[entry.Name()] = string(componentTemplateString)
		}
	}

	// Read each node template
	entries, err = templatesFS.ReadDir(NodeTemplatesPath)
	if err != nil {
		return fmt.Errorf("failed to read node templates directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			nodeTemplatePath := filepath.Join(NodeTemplatesPath, entry.Name())
			nodeTemplateString, err := templatesFS.ReadFile(nodeTemplatePath)
			if err != nil {
				return fmt.Errorf("failed to read node template %s: %w", entry.Name(), err)
			}
			nodeTemplates[entry.Name()] = string(nodeTemplateString)
		}
	}

	return nil
}

// Return the error template
func (s *templatesService) GetErrorTemplate(tenant, realm, flowId string) (*template.Template, error) {

	// TODO this should just be an error node
	return s.GetTemplates(tenant, realm, flowId, "error")
}
