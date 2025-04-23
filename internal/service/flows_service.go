package service

import (
	"fmt"
	"goiam/internal/config"
	"goiam/internal/logger"
	"goiam/internal/model"
	"os"
	"path/filepath"
	"sync"
)

// FlowService defines the business logic for flow operations
type FlowService interface {

	// Initialize Flows
	InitFlows() error

	// GetFlow returns a flow by its ID
	GetFlowById(tenant, realm, id string) (*model.FlowWithRoute, bool)

	// GetFlowByPath returns a flow by its path
	GetFlowByPath(tenant, realm, path string) (*model.FlowWithRoute, bool)

	// ListFlows returns all flows
	ListFlows(tenant, realm string) ([]*model.FlowWithRoute, error)

	// ListAllFlows returns all flows or all realms
	ListAllFlows() ([]*model.FlowWithRoute, error)

	// CreateFlow creates a new flow
	CreateFlow(tenant, realm string, flow *model.FlowWithRoute) error
}

// flowServiceImpl implements FlowService
type flowServiceImpl struct {

	// flows is a map of flow name to flow with route, flow name is used as key
	flows   map[string]*model.FlowWithRoute
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

func (s *flowServiceImpl) GetFlowById(tenant, realm, id string) (*model.FlowWithRoute, bool) {

	// TODO we need to filter the flows by tenant and realm

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	flow, ok := s.flows[id]
	if !ok {
		return nil, false
	}

	return flow, true
}

func (s *flowServiceImpl) GetFlowByPath(tenant, realm, path string) (*model.FlowWithRoute, bool) {

	// TODO we need to filter the flows by tenant and realm

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	// For the time being we just iterate over all flows and return the first one that matches the path
	for _, flow := range s.flows {
		if flow.Route == "/"+path {
			return flow, true
		}
	}

	return nil, false
}

func (s *flowServiceImpl) ListFlows(tenant, realm string) ([]*model.FlowWithRoute, error) {

	// TODO we need to filter the flows by tenant and realm

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	flows := make([]*model.FlowWithRoute, 0, len(s.flows))
	for _, flow := range s.flows {
		flows = append(flows, flow)
	}

	return flows, nil
}

func (s *flowServiceImpl) ListAllFlows() ([]*model.FlowWithRoute, error) {

	s.flowsMu.RLock()
	defer s.flowsMu.RUnlock()

	// Return a copy of the flows
	flows := make([]*model.FlowWithRoute, 0, len(s.flows))
	for _, flow := range s.flows {
		flows = append(flows, flow)
	}

	return flows, nil
}

func (s *flowServiceImpl) CreateFlow(tenant, realm string, flow *model.FlowWithRoute) error {

	s.flowsMu.Lock()
	defer s.flowsMu.Unlock()

	s.flows[flow.Flow.Name] = flow

	return nil
}

func (s *flowServiceImpl) initFlowsFromConfigDir(configRoot string) error {

	logger.DebugNoContext("Initializing flows from config dir: %s", configRoot)

	s.flowsMu.Lock()
	defer s.flowsMu.Unlock()

	// clear the map
	s.flows = make(map[string]*model.FlowWithRoute)

	// Get the flows directory
	flowsDir := filepath.Join(configRoot, "flows")

	// check if the dir exists
	if _, err := os.Stat(flowsDir); os.IsNotExist(err) {
		return fmt.Errorf("flows directory %s does not exist", flowsDir)
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
		if _, ok := s.flows[flow.Flow.Name]; ok {
			return fmt.Errorf("flow name %s already in use", flow.Flow.Name)
		}

		// Add the flow to the map with the flow name as key
		s.flows[flow.Flow.Name] = flow
	}

	return nil
}

/*
func (s *realmServiceImpl) LookupFlow(tenant, realm, path string) (*model.FlowWithRoute, error) {
	realmID := fmt.Sprintf("%s/%s", tenant, realm)
	loaded, ok := s.GetRealm(realmID)
	if !ok {
		return nil, errors.New("realm not found")
	}

	cleanPath := "/" + strings.TrimPrefix(path, "/")

	for _, f := range loaded.Config.Flows {
		if f.Route == cleanPath {
			return &f, nil
		}
	}

	return nil, errors.New("route not found in realm")
}

func (s *realmServiceImpl) ListFlowsPerRealm(tenant, realm string) ([]model.FlowWithRoute, error) {
	realmID := fmt.Sprintf("%s/%s", tenant, realm)
	loaded, ok := s.GetRealm(realmID)
	if !ok {
		return nil, fmt.Errorf("realm not found: %s", realmID)
	}
	return loaded.Config.Flows, nil
}

func (s *realmServiceImpl) LookupFlowByName(tenant, realm, name string) (*model.FlowWithRoute, error) {
	flows, err := s.ListFlowsPerRealm(tenant, realm)
	if err != nil {
		return nil, err
	}

	for _, f := range flows {
		if f.Flow != nil && f.Flow.Name == name {
			return &f, nil
		}
	}

	return nil, errors.New("flow not found: " + name)
}*/
