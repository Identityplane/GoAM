package graph

import (
	"errors"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
)

// Enum of node types
type NodeType string

const (
	NodeTypeInit           NodeType = "init"
	NodeTypeQuery          NodeType = "query"
	NodeTypeQueryWithLogic NodeType = "queryWithLogic"
	NodeTypeLogic          NodeType = "logic"
	NodeTypeResult         NodeType = "result"
)

type NodeDefinition struct {
	Name            string                                                                                                           `json:"name"`       // e.g. "askUsername"
	Type            NodeType                                                                                                         `json:"type"`       // query, logic, etc.
	RequiredContext []string                                                                                                         `json:"inputs"`     // field that the node requires from the flow context
	OutputContext   []string                                                                                                         `json:"outputs"`    // fields that the node will set in the flow context
	Prompts         map[string]string                                                                                                `json:"prompts"`    // key: label/type shown to user, will be returned via the user input argument
	Conditions      []string                                                                                                         `json:"conditions"` // e.g. ["success", "fail"]
	Run             func(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) // Run function for logic nodes, must either return a condition or a set of prompts
}

type NodeResult struct {
	Prompts   map[string]string // Prompts to be shown to the user, if applicable
	Condition string            // The next state, if applicable
}

type GraphNode struct {
	Name         string            `json:"name"`                    // unique in graph
	Use          string            `json:"use"`                     // reference to NodeDefinition.Name
	Next         map[string]string `json:"next"`                    // condition -> next GraphNode.Name
	CustomConfig map[string]string `json:"custom_config,omitempty"` // for overrides (optional)
}

type FlowDefinition struct {
	Name  string                `json:"name"`
	Start string                `json:"start"` // e.g., "init"
	Nodes map[string]*GraphNode `json:"nodes"`
}

type FlowState struct {
	RunID   string            `json:"run_id"`
	Current string            `json:"current"` // active node
	Context map[string]string `json:"context"` // dynamic values (inputs + outputs)
	History []string          `json:"history"` // executed node names
	Error   *string           `json:"error,omitempty"`
	Result  *FlowResult       `json:"result,omitempty"`
	User    *model.User       `json:"user,omitempty"`
	Prompts map[string]string `json:"prompts,omitempty"` // Prompts to be shown to the user, if applicable
}

type AuthLevel string

const (
	AuthLevelUnauthenticated AuthLevel = "0"
	AuthLevel1FA             AuthLevel = "1"
	AuthLevel2FA             AuthLevel = "2"
)

type FlowResult struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Authenticated bool      `json:"authenticated"`
	AuthLevel     AuthLevel `json:"auth_level"`
	FlowName      string    `json:"flow_name,omitempty"`
}

type ServiceRegistry struct {
	UserRepo repository.UserRepository
}

// Create NodeResult with state
func NewNodeResultWithCondition(condition string) (*NodeResult, error) {
	return &NodeResult{
		Prompts:   nil,
		Condition: condition,
	}, nil
}

// Create NodeResult with prompts
func NewNodeResultWithPrompts(prompts map[string]string) (*NodeResult, error) {
	return &NodeResult{
		Prompts:   prompts,
		Condition: "",
	}, nil
}

// Create NodeResult with error
func NewNodeResultWithError(err error) (*NodeResult, error) {
	return nil, err
}

// Create NodeResult with text error
func NewNodeResultWithTextError(text string) (*NodeResult, error) {
	return nil, errors.New(text)
}
