package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/graph"
	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

const DEFAULT_FLOW_DEFINITION = `description: 'An empty flow'
start: init
nodes:
  init:
    name: init
    use: init
    next:
      start: failure
  failure:
    name: failureResult
    use: failureResult
    next: {}
  sucess:
    name: successResult
    use: successResult
    next: {}
editor:
  nodes:
    init:
      x: 0
      'y': 200
    node_75e21c1d:
      x: 300
      'y': 300
    node_8840e55b:
      x: 300
      'y': 200
`

// flowServiceImpl implements FlowService
type flowServiceImpl struct {
	flowsDb db.FlowDB
}

// NewFlowService creates a new FlowService instance
func NewFlowService(flowsDb db.FlowDB) services_interface.FlowService {
	return &flowServiceImpl{
		flowsDb: flowsDb,
	}
}

func (s *flowServiceImpl) GetFlowById(tenant, realm, id string) (*model.Flow, bool) {
	flow, err := s.flowsDb.GetFlow(context.Background(), tenant, realm, id)
	if err != nil || flow == nil {
		return nil, false
	}

	// load flow defenition from yaml if yaml is not ""
	if flow.DefinitionYaml != "" {

		flow.Definition, err = lib.LoadFlowDefinitonFromString(flow.DefinitionYaml)

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
	if flow.DefinitionYaml != "" {

		flow.Definition, err = lib.LoadFlowDefinitonFromString(flow.DefinitionYaml)

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

	// If the flow definition is yet, we validate it
	if flow.DefinitionYaml != "" {
		lintErrors, err := s.ValidateFlowDefinition(flow.DefinitionYaml)
		if err != nil {
			return fmt.Errorf("failed to validate flow definition: %w", err)
		}
		if len(lintErrors) > 0 {
			return fmt.Errorf("flow definition is invalid: %v", lintErrors)
		}
	} else {
		// If the flow definition is not set, we set it to an default flow definition
		flow.DefinitionYaml = DEFAULT_FLOW_DEFINITION
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

func (s *flowServiceImpl) ValidateFlowDefinition(content string) ([]services_interface.FlowLintError, error) {
	// Try to parse the YAML content
	flowDefinition, err := lib.LoadFlowDefinitonFromString(content)

	if err != nil {
		return []services_interface.FlowLintError{{
			StartLineNumber: 1,
			StartColumn:     1,
			EndLineNumber:   1,
			EndColumn:       1,
			Message:         fmt.Sprintf("%s", err.Error()),
			Severity:        8,
		}}, nil
	}

	error := graph.ValidateFlowDefinition(flowDefinition)

	if error != nil {
		return []services_interface.FlowLintError{{
			StartLineNumber: 1,
			StartColumn:     1,
			EndLineNumber:   1,
			EndColumn:       1,
			Message:         fmt.Sprintf("%s", error.Error()),
			Severity:        8,
		}}, nil
	}

	return []services_interface.FlowLintError{}, nil
}
