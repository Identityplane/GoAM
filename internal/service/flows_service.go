package service

import (
	"context"
	"fmt"
	"goiam/internal/auth/graph"
	"goiam/internal/config"
	"goiam/internal/db"
	"goiam/internal/logger"
	"goiam/internal/model"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FlowService defines the business logic for flow operations
type FlowService interface {
	// Initialize Flows
	InitFlows() error

	// GetFlow returns a flow by its ID
	GetFlowById(tenant, realm, id string) (*model.Flow, bool)

	// GetFlowByPath returns a flow by its path
	GetFlowByPath(tenant, realm, path string) (*model.Flow, bool)

	// ListFlows returns all flows
	ListFlows(tenant, realm string) ([]model.Flow, error)

	// ListAllFlows returns all flows for all realms
	ListAllFlows() ([]model.Flow, error)

	// CreateFlow creates a new flow
	CreateFlow(tenant, realm string, flow model.Flow) error

	// UpdateFlow updates an existing flow
	UpdateFlow(tenant, realm string, flow model.Flow) error

	// DeleteFlow deletes a flow by its ID
	DeleteFlow(tenant, realm, id string) error

	// ValidateFlowDefinition validates a YAML flow definition
	ValidateFlowDefinition(content string) ([]FlowLintError, error)
}

// flowServiceImpl implements FlowService
type flowServiceImpl struct {
	flowsDb db.FlowDB
}

// NewFlowService creates a new FlowService instance
func NewFlowService(flowsDb db.FlowDB) FlowService {
	return &flowServiceImpl{
		flowsDb: flowsDb,
	}
}

func (s *flowServiceImpl) InitFlows() error {
	err := s.initFlowsFromConfigDir(config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to init flows from config dir: %w", err)
	}

	return nil
}

func (s *flowServiceImpl) GetFlowById(tenant, realm, id string) (*model.Flow, bool) {
	flow, err := s.flowsDb.GetFlow(context.Background(), tenant, realm, id)
	if err != nil || flow == nil {
		return nil, false
	}

	// load flow defenition from yaml if yaml is not ""
	if flow.DefintionYaml != "" {

		flow.Definition, err = LoadFlowDefinitonFromString(flow.DefintionYaml)

		if err != nil {
			return nil, false
		}
	}

	return flow, true
}

func (s *flowServiceImpl) GetFlowByPath(tenant, realm, path string) (*model.Flow, bool) {
	flow, err := s.flowsDb.GetFlowByRoute(context.Background(), tenant, realm, path)
	if err != nil || flow == nil {
		return nil, false
	}

	// load flow defenition from yaml if yaml is not ""
	if flow.DefintionYaml != "" {

		flow.Definition, err = LoadFlowDefinitonFromString(flow.DefintionYaml)

		if err != nil {
			return nil, false
		}
	}

	return flow, true
}

func (s *flowServiceImpl) ListFlows(tenant, realm string) ([]model.Flow, error) {
	flows, err := s.flowsDb.ListFlows(context.Background(), tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list flows: %w", err)
	}
	return flows, nil
}

func (s *flowServiceImpl) ListAllFlows() ([]model.Flow, error) {
	flows, err := s.flowsDb.ListAllFlows(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list all flows: %w", err)
	}
	return flows, nil
}

func (s *flowServiceImpl) CreateFlow(tenant, realm string, flow model.Flow) error {
	// Check that route is not ""
	if flow.Route == "" {
		return fmt.Errorf("flow route is empty")
	}

	// Check that flow id is not ""
	if flow.Id == "" {
		return fmt.Errorf("flow id is empty")
	}

	// Ensure realm and tenant are set correctly
	flow.Realm = realm
	flow.Tenant = tenant

	// We ignore heading "/" for the route name
	flow.Route, _ = strings.CutPrefix(flow.Route, "/")

	// Check if the flow already exists
	_, exists := s.GetFlowById(tenant, realm, flow.Id)
	if exists {
		return fmt.Errorf("flow with id %s already exists", flow.Id)
	}

	// Create the flow in the database
	return s.flowsDb.CreateFlow(context.Background(), flow)
}

func (s *flowServiceImpl) UpdateFlow(tenant, realm string, flow model.Flow) error {

	// Check that route is not ""
	if flow.Route == "" {
		return fmt.Errorf("flow route is empty")
	}

	// Check that flow id is not ""
	if flow.Id == "" {
		return fmt.Errorf("flow id is empty")
	}

	// Ensure realm and tenant are set correctly
	flow.Realm = realm
	flow.Tenant = tenant

	// We ignore heading "/" for the route name
	flow.Route, _ = strings.CutPrefix(flow.Route, "/")

	// Check if the flow exists
	_, exists := s.GetFlowById(tenant, realm, flow.Id)
	if !exists {
		return fmt.Errorf("flow with id %s not found", flow.Id)
	}

	// Update the flow in the database
	return s.flowsDb.UpdateFlow(context.Background(), &flow)
}

func (s *flowServiceImpl) DeleteFlow(tenant, realm, id string) error {

	// Get the flow first to check if it exists
	_, exists := s.GetFlowById(tenant, realm, id)
	if !exists {
		return fmt.Errorf("flow with id %s not found", id)
	}

	// Delete the flow from the database
	return s.flowsDb.DeleteFlow(context.Background(), tenant, realm, id)
}

func (s *flowServiceImpl) initFlowsFromConfigDir(configRoot string) error {
	logger.DebugNoContext("Initializing flows from config dir: %s", configRoot)

	// Load all flows for all realms
	allRealms, err := services.RealmService.GetAllRealms()
	if err != nil {
		return fmt.Errorf("failed to load all realms while initFlowsFromConfigDir: %s", err)
	}

	for _, realm := range allRealms {
		err := s.loadFlowsFromRealmConfigDir(realm.Config.Tenant, realm.Config.Realm, configRoot)
		if err != nil {
			return fmt.Errorf("failed to load flows from realm %s: %w", realm.RealmID, err)
		}
	}

	return nil
}

func (s *flowServiceImpl) loadFlowsFromRealmConfigDir(tenant, realm, configRoot string) error {
	flowsDir := filepath.Join(configRoot, "tenants", tenant, realm, "flows")

	// check if the dir exists
	if _, err := os.Stat(flowsDir); os.IsNotExist(err) {
		logger.ErrorNoContext("flows directory %s does not exist", flowsDir)
		return nil
	}

	// Go over all yaml files in the flows directory
	files, err := filepath.Glob(filepath.Join(flowsDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list flows in %s: %w", flowsDir, err)
	}

	// Load each flow
	for _, file := range files {
		flow, err := LoadFlowFromYAML(file)
		if err != nil {
			return fmt.Errorf("failed to load flow from %s: %w", file, err)
		}

		// Check if the flow already exists
		_, exists := s.GetFlowById(tenant, realm, flow.Id)
		if exists {
			logger.InfoNoContext("flow %s already exists, skipping from config file", flow.Id)
			continue
		}

		// Create the flow in the database
		err = s.CreateFlow(tenant, realm, *flow)
		if err != nil {
			return fmt.Errorf("failed to create flow %s: %w", flow.Id, err)
		}
	}

	return nil
}

type FlowLintError struct {
	StartLineNumber int    `json:"startLineNumber"`
	StartColumn     int    `json:"startColumn"`
	EndLineNumber   int    `json:"endLineNumber"`
	EndColumn       int    `json:"endColumn"`
	Message         string `json:"message"`
	Severity        int    `json:"severity"`
}

func (s *flowServiceImpl) ValidateFlowDefinition(content string) ([]FlowLintError, error) {
	// Try to parse the YAML content
	var yflow yamlFlowDefinition
	if err := yaml.Unmarshal([]byte(content), &yflow); err != nil {
		return []FlowLintError{{
			StartLineNumber: 1,
			StartColumn:     1,
			EndLineNumber:   1,
			EndColumn:       1,
			Message:         fmt.Sprintf("Invalid YAML format: %v", err),
			Severity:        8,
		}}, nil
	}

	flowDefinition := yflow.ConvertToFlowDefinition()
	error := graph.ValidateFlowDefinition(flowDefinition)

	if error != nil {
		return []FlowLintError{{
			StartLineNumber: 1,
			StartColumn:     1,
			EndLineNumber:   1,
			EndColumn:       1,
			Message:         fmt.Sprintf("%s", error.Error()),
			Severity:        8,
		}}, nil
	}

	return []FlowLintError{}, nil
}
