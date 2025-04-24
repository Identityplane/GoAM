package service

import (
	"fmt"
	"goiam/internal/config"
	"goiam/internal/logger"
	"goiam/internal/model"
	"os"
	"path/filepath"
	"strings"
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

	// For now we just go over all flows and return the first one that matches the id and realm
	for _, flow := range s.flows {
		if flow.Realm == realm && flow.Tenant == tenant && flow.Flow.Name == id {
			return flow, true
		}
	}

	return nil, false
}

func (s *flowServiceImpl) GetFlowByPath(tenant, realm, path string) (*model.FlowWithRoute, bool) {

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

	// We ignore heading "/" for the route name
	flow.Route, _ = strings.CutPrefix(flow.Route, "/")

	// Ensure realm and tenant are set correctly
	flow.Realm = realm
	flow.Tenant = tenant

	// Store flow in memory
	s.flows[tenant+"/"+realm+"/"+flow.Route] = flow

	return nil
}

func (s *flowServiceImpl) initFlowsFromConfigDir(configRoot string) error {

	logger.DebugNoContext("Initializing flows from config dir: %s", configRoot)

	// clear the map
	s.flows = make(map[string]*model.FlowWithRoute)

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
		logger.DebugNoContext("loaded flow %s from %s for realmId %s", flow.Flow.Name, file, tenant+"/"+realm)

		s.CreateFlow(tenant, realm, flow)
	}

	return nil
}
