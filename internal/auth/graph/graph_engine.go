package graph

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Engine struct {
	Flow *FlowDefinition
}

// NewEngine constructs and validates a flow
func NewEngine(def *FlowDefinition) (*Engine, error) {
	engine := &Engine{
		Flow: def,
	}
	if err := engine.validate(); err != nil {
		return nil, err
	}

	return engine, nil
}

// validate checks for basic structural integrity of the flow
func (e *Engine) validate() error {

	// Check if start node is in map
	_, ok := e.Flow.Nodes[e.Flow.Start]
	if !ok {
		return fmt.Errorf("start node '%s' not found in nodes", e.Flow.Start)
	}
	def := e.Flow

	// Check start node is of type 'init'
	if def.Start == "" {
		return errors.New("flow start node is not defined")
	}
	if def.Nodes[def.Start] == nil {
		return fmt.Errorf("start node '%s' is missing from the graph", def.Start)
	}
	if nodeDef := def.Nodes[def.Start]; nodeDef.Use != "init" {
		return fmt.Errorf("start node '%s' must be of type 'init'", def.Start)
	}

	// Check non-terminal nodes have a Next map
	for name, node := range def.Nodes {
		nodeType := NodeTypeInit
		if def := getNodeDefinitionByName(node.Use); def != nil {
			nodeType = def.Type
		} else if node.Use == "successResult" || node.Use == "failureResult" {
			nodeType = NodeTypeResult
		}

		if nodeType != NodeTypeResult && node.Next == nil {
			return fmt.Errorf("node '%s' must define a 'Next' map", name)
		}
	}
	return nil
}

// InitFlow creates a new FlowState for a given flow
func InitFlow(flow *FlowDefinition) *FlowState {
	return &FlowState{
		RunID:   uuid.NewString(),
		Current: flow.Start,
		Context: make(map[string]string),
		History: []string{},
	}
}

// Run processes one step of the flow and returns either:
// - prompts for query node
// - result node
// - or nil and continues internally
func Run(flow *FlowDefinition, state *FlowState, inputs map[string]string) (map[string]string, *GraphNode, error) {
	if state == nil || state.Current == "" {
		return nil, nil, errors.New("invalid or uninitialized flow state")
	}

	node, ok := flow.Nodes[state.Current]
	if !ok {
		return nil, nil, fmt.Errorf("node '%s' not found in flow", state.Current)
	}

	def := getNodeDefinitionByName(node.Use)
	if def == nil {
		return nil, nil, fmt.Errorf("node definition for '%s' not found", node.Use)
	}

	switch def.Type {
	case NodeTypeInit, NodeTypeLogic:
		condition, err := runNodeLogic(state, node)
		if err != nil {
			msg := err.Error()
			state.Error = &msg
			return nil, nil, err
		}

		// Add to history with condition
		state.History = append(state.History, fmt.Sprintf("%s:%s", node.Name, condition))

		next, ok := node.Next[condition]
		if !ok {
			return nil, nil, fmt.Errorf("no transition defined for condition '%s' in node '%s'", condition, node.Name)
		}
		state.Current = next
		return Run(flow, state, nil)

	case NodeTypeQuery:
		if inputs == nil {
			state.History = append(state.History, state.Current)
			return def.Prompts, nil, nil
		}

		inputSummary := make(map[string]string)
		for inputKey := range def.Prompts {
			val, ok := inputs[inputKey]
			if !ok || val == "" {
				return nil, nil, fmt.Errorf("missing value for prompt '%s'", inputKey)
			}
			state.Context[inputKey] = val
			inputSummary[inputKey] = val
		}

		// Marshal inputs for history
		inputJson, err := json.Marshal(inputSummary)
		if err != nil {
			inputJson = []byte(`{}`)
		}
		state.History = append(state.History, fmt.Sprintf("%s:submitted:%s", node.Name, inputJson))

		state.Current = node.Next["submitted"]
		return Run(flow, state, nil)

	case NodeTypeResult:
		state.History = append(state.History, state.Current)
		return nil, node, nil

	case NodeTypeQueryWithLogic:

		// CASE Answer
		// TODO Implement case where answer is sent

		// CASE Promt
		nodeDefinition := NodeDefinitions[node.Name]

		// generate the prompts
		thePrompts, err := nodeDefinition.GeneratePrompts(state, node)
		if err != nil {
			return nil, nil, err
		}

		state.History = append(state.History, state.Current)
		return thePrompts, nil, nil

	default:
		return nil, nil, fmt.Errorf("unsupported node type: %s", def.Type)
	}
}

// runNodeLogic runs built-in logic for known nodes
func runNodeLogic(state *FlowState, node *GraphNode) (string, error) {

	logicFunc, ok := LogicFunctions[node.Use]
	if !ok {
		return "", fmt.Errorf("no logic function registered for node '%s'", node.Use)
	}

	// Execute node
	condition, err := logicFunc(state, node)
	if err != nil {
		return "", err
	}

	// Check if return conditon is valid
	def := getNodeDefinitionByName(node.Use)
	if def == nil {
		return "", fmt.Errorf("no node definition found for '%s'", node.Use)
	}

	valid := false
	for _, c := range def.Conditions {
		if c == condition {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid condition '%s' returned by logic for node '%s'", condition, node.Name)
	}

	return condition, nil

}

func getNodeDefinitionByName(name string) *NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
