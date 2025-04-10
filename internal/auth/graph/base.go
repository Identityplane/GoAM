package graph

import (
	"goiam/internal/auth/repository"
	"goiam/internal/db/model"
)

// Enum of node types
type NodeType string

const (
	NodeTypeInit   NodeType = "init"
	NodeTypeQuery  NodeType = "query"
	NodeTypeLogic  NodeType = "logic"
	NodeTypeResult NodeType = "result"
)

type NodeDefinition struct {
	Name       string            `json:"name"`       // e.g. "askUsername"
	Type       NodeType          `json:"type"`       // query, logic, etc.
	Inputs     []string          `json:"inputs"`     // required context fields
	Outputs    []string          `json:"outputs"`    // new fields it will write
	Prompts    map[string]string `json:"prompts"`    // key: label/type shown to user
	Conditions []string          `json:"conditions"` // e.g. ["success", "fail"]
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
}

type LogicFunc func(state *FlowState, node *GraphNode) (condition string, err error)

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

var Services = &ServiceRegistry{}

type ServiceRegistry struct {
	UserRepo repository.UserRepository
}
