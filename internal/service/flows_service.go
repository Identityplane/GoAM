package service

import (
	"fmt"
	"goiam/internal/auth/graph"
	"goiam/internal/config"
	"goiam/internal/logger"
	"goiam/internal/model"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	ListFlows(tenant, realm string) ([]*model.Flow, error)

	// ListAllFlows returns all flows or all realms
	ListAllFlows() ([]*model.Flow, error)

	// CreateFlow creates a new flow
	CreateFlow(tenant, realm string, flow *model.Flow) error

	// DeleteFlow deletes a flow by its ID
	DeleteFlow(tenant, realm, id string) error

	// ValidateFlowDefinition validates a YAML flow definition
	ValidateFlowDefinition(content string) ([]FlowLintError, error)
}

// flowServiceImpl implements FlowService
type flowServiceImpl struct {

	// flows is a map of flow name to flow with route, flow name is used as key
	flows   map[string]*model.Flow
	flowsMu sync.RWMutex
}

// NewFlowService creates a new FlowService instance
func NewFlowService() FlowService {
	return &flowServiceImpl{}
}

func (s *flowServiceImpl) InitFlows() error {

	err := s.initFlowsFromConfigDir(config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to init flows from config dir: %w", err)
	}

	return nil
}

func (s *flowServiceImpl) GetFlowById(tenant, realm, id string) (*model.Flow, bool) {

	// TODO we need to filter the flows by tenant and realm

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	// For now we just go over all flows and return the first one that matches the id and realm
	for _, flow := range s.flows {
		if flow.Realm == realm && flow.Tenant == tenant && flow.Id == id {
			return flow, true
		}
	}

	return nil, false
}

func (s *flowServiceImpl) GetFlowByPath(tenant, realm, path string) (*model.Flow, bool) {

	// TODO we need to filter the flows by tenant and realm

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	// For the time being we just iterate over all flows and return the first one that matches the path
	flow, ok := s.flows[tenant+"/"+realm+"/"+path]
	if !ok {
		return nil, false
	}

	return flow, true
}

func (s *flowServiceImpl) ListFlows(tenant, realm string) ([]*model.Flow, error) {

	// TODO we need to filter the flows by tenant and realm

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	flows := make([]*model.Flow, 0, len(s.flows))
	for _, flow := range s.flows {
		flows = append(flows, flow)
	}

	return flows, nil
}

func (s *flowServiceImpl) ListAllFlows() ([]*model.Flow, error) {

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	// Return a copy of the flows
	flows := make([]*model.Flow, 0, len(s.flows))
	for _, flow := range s.flows {
		flows = append(flows, flow)
	}

	return flows, nil
}

func (s *flowServiceImpl) CreateFlow(tenant, realm string, flow *model.Flow) error {

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

	// check if flow already exisits by query for id, if it already exists we need to delete it first
	_, exists := s.GetFlowById(tenant, realm, flow.Id)
	if exists {
		err := s.DeleteFlow(tenant, realm, flow.Id)
		if err != nil {
			return fmt.Errorf("failed to delete flow %s: %w", flow.Id, err)
		}
	}

	s.flowsMu.Lock()
	defer s.flowsMu.Unlock()

	// We ignore heading "/" for the route name
	flow.Route, _ = strings.CutPrefix(flow.Route, "/")

	// Store flow in memory
	s.flows[tenant+"/"+realm+"/"+flow.Route] = flow

	return nil
}

func (s *flowServiceImpl) DeleteFlow(tenant, realm, id string) error {
	// Get the flow first to check if it exists and get its route
	flow, exists := s.GetFlowById(tenant, realm, id)
	if !exists {
		return fmt.Errorf("flow with id %s not found", id)
	}

	s.flowsMu.Lock()
	defer s.flowsMu.Unlock()

	// Delete the flow from the map using the tenant/realm/route key
	delete(s.flows, tenant+"/"+realm+"/"+flow.Route)

	return nil
}

func (s *flowServiceImpl) initFlowsFromConfigDir(configRoot string) error {

	logger.DebugNoContext("Initializing flows from config dir: %s", configRoot)

	// clear the map
	s.flows = make(map[string]*model.Flow)

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

		// check if flow name is already in the map
		if _, ok := s.flows[tenant+"/"+realm+"/"+flow.Route]; ok {
			return fmt.Errorf("flow name %s already in use", tenant+"/"+realm+"/"+flow.Route)
		}

		// Add the flow to the map with the flow name as key
		logger.DebugNoContext("loaded flow %s from %s for realmId %s", flow.Id, file, tenant+"/"+realm)

		err = s.CreateFlow(tenant, realm, flow)
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
			Message:         fmt.Sprintf(error.Error()),
			Severity:        8,
		}}, nil
	}

	return []FlowLintError{}, nil
}
