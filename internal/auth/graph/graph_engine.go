package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

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

// InitFlow creates a new FlowState for a given flow
func InitFlow(flow *FlowDefinition) *FlowState {
	return &FlowState{
		RunID:   uuid.NewString(),
		Current: flow.Start,
		Context: make(map[string]string),
		History: []string{},
	}
}

// Run processes one step of the flow and returns either
// - flow state when graph is requesting prompt or is finished
// - error if any internal error occurred
func Run(flow *FlowDefinition, state *FlowState, inputs map[string]string) (*FlowState, error) {

	// Check if state is present and valid
	if state == nil || state.Current == "" {
		return nil, errors.New("invalid or uninitialized flow state")
	}

	// Check if node for current state exists in flow
	node, ok := flow.Nodes[state.Current]
	if !ok {
		return nil, fmt.Errorf("node '%s' not found in flow", state.Current)
	}

	// Load node definition from node name
	def := getNodeDefinitionByName(node.Use)
	if def == nil {
		return nil, fmt.Errorf("node definition for '%s' not found", node.Use)
	}

	var nodeResult *NodeResult
	var err error

	// Process node by type
	switch def.Type {
	case NodeTypeInit:
		nodeResult, err = ProcessInitTypeNode(state, node, def, inputs)

	case NodeTypeLogic:
		nodeResult, err = ProcessLogicTypeNode(state, node, def, inputs)

	case NodeTypeQuery:
		nodeResult, err = ProcessQueryTypeNode(state, node, def, inputs)

	case NodeTypeResult:
		nodeResult, err = ProcessResultTypeNode(state, node, def, inputs)

	case NodeTypeQueryWithLogic:
		nodeResult, err = ProcessQueryWithLogicTypeNode(state, node, def, inputs)

	default:
		return nil, fmt.Errorf("unsupported node type: %s", def.Type)
	}

	// Return error if present
	if err != nil {
		log.Printf("Error processing node '%s': %v", node.Name, err)
		return nil, err
	}

	// End the graph if the node is a result node
	if def.Type == NodeTypeResult {

		return state, nil
	}

	// if there are prompt in the result we update the state and return
	if nodeResult.Prompts != nil {

		// turn the nodeResult.Prompts into a strong for logging
		promptsString, err := json.Marshal(nodeResult.Prompts)
		if err != nil {
			log.Printf("Error marshalling prompts: %v", err)
			return nil, err
		}

		// log the node name, type and prompts
		log.Printf("Node %s of type %s resulted in prompts %s", node.Name, def.Type, promptsString)
		state.History = append(state.History, fmt.Sprintf("%s:prompted:%s", node.Name, promptsString))

		// Update prompts in string and return
		state.Prompts = nodeResult.Prompts
		return state, nil
	}
	if nodeResult.Condition != "" {

		// log the node name and condition
		condition := nodeResult.Condition
		state.History = append(state.History, fmt.Sprintf("%s:%s", node.Name, condition))

		// Check if resulting condition is valid as defined in the node definiton
		valid := false
		for _, c := range def.Conditions {
			if c == condition {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid condition '%s' returned by node '%s'", condition, node.Name)
		}

		// Clear prompts if no prompts are present
		state.Prompts = nil

		// lookup transition in graph
		if nextNodeName, ok := node.Next[condition]; ok {
			state.Current = nextNodeName
		} else {
			return nil, fmt.Errorf("no next node defined for condition '%s'", condition)
		}

		// Clear inputs as next node does not expect any inputs
		inputs = nil

		// recursively call run with the new state
		return Run(flow, state, inputs)
	}

	// throw an error if no condition or prompts are present
	panic(fmt.Sprintf("node '%s' returned neither prompts nor condition", node.Name))
}

// ProcessQueryTypeNode processes a query node
// and returns the next state and any prompts to be shown to the user
func ProcessQueryTypeNode(state *FlowState, node *GraphNode, def *NodeDefinition, inputs map[string]string) (*NodeResult, error) {

	// If no inputs are present send prompts to user
	if inputs == nil {

		return &NodeResult{Prompts: def.Prompts, Condition: ""}, nil
	}

	// Else if we have inputs to context and return submitted
	for k, v := range inputs {
		state.Context[k] = v
	}
	return &NodeResult{Prompts: nil, Condition: "submitted"}, nil
}

func ProcessResultTypeNode(state *FlowState, node *GraphNode, def *NodeDefinition, inputs map[string]string) (*NodeResult, error) {

	// we expect no inputs for a result node
	if inputs != nil {
		return nil, fmt.Errorf("result node '%s' must not have inputs", node.Name)
	}

	// run the node logic
	// Check if den.Run is not nil
	if def.Run == nil {
		return nil, fmt.Errorf("result node '%s' has no run function", node.Name)
	}

	result, err := def.Run(state, node, nil)

	// we expect the result to have no prompts and no condition as this is a terminal node
	if err != nil {
		return nil, err
	}
	if result.Prompts != nil {
		return nil, fmt.Errorf("result node '%s' must not have prompts", node.Name)
	}
	if result.Condition != "" {
		return nil, fmt.Errorf("result node '%s' must not have condition", node.Name)
	}

	// The result node must set the flow result
	if state.Result == nil {
		return nil, fmt.Errorf("result node '%s' must set the flow result", node.Name)
	}

	// update history
	state.History = append(state.History, node.Name)

	return result, nil
}

// ProcessInitTypeNode processes an init node
// and returns the next state and any prompts to be shown to the user
func ProcessInitTypeNode(state *FlowState, node *GraphNode, def *NodeDefinition, inputs map[string]string) (*NodeResult, error) {

	// Run init node logic
	result, err := def.Run(state, node, inputs)

	if err != nil {
		return nil, err
	}

	// init node must return a condition
	if result.Condition == "" {
		return nil, fmt.Errorf("init node '%s' must return a condition", node.Name)
	}

	return result, nil
}

// ProcessLogicTypeNode processes a logic node
// and returns the next state
func ProcessLogicTypeNode(state *FlowState, node *GraphNode, def *NodeDefinition, inputs map[string]string) (*NodeResult, error) {

	// Run node logic
	result, err := def.Run(state, node, inputs)
	if err != nil {
		return nil, err
	}
	// Check if result is valid
	if result.Condition == "" {
		return nil, fmt.Errorf("logic node '%s' must return a condition", node.Name)
	}

	return result, nil
}

// Process NodeTypeQueryWithLogic node
// and returns the next state and any prompts to be shown to the user
func ProcessQueryWithLogicTypeNode(state *FlowState, node *GraphNode, def *NodeDefinition, inputs map[string]string) (*NodeResult, error) {

	// Run node logic
	result, err := def.Run(state, node, inputs)

	if err != nil {
		return nil, err
	}

	// check if the result is a prompt or a condition
	if result.Prompts != nil {

		return result, nil
	} else if result.Condition != "" {

		return result, nil
	}

	// if no result is returned, return an error
	return nil, fmt.Errorf("query node '%s' must return a prompt or a condition", node.Name)
}

func getNodeDefinitionByName(name string) *NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
