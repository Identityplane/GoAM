package graph

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
)

var log = logger.GetGoamLogger()

const MAX_HISTORY_SIZE = 100

type Engine struct {
	Flow *model.FlowDefinition
}

// NewEngine constructs and validates a flow
func NewEngine(def *model.FlowDefinition) (*Engine, error) {
	engine := &Engine{
		Flow: def,
	}
	if err := ValidateFlowDefinition(def); err != nil {
		return nil, err
	}

	return engine, nil
}

// InitFlow creates a new FlowState for a given flow
func InitFlow(flow *model.FlowDefinition) *model.AuthenticationSession {
	return &model.AuthenticationSession{
		RunID:   uuid.NewString(),
		Current: flow.Start,
		Context: make(map[string]string),
		History: []string{},
	}
}

// Run processes one step of the flow and returns either
// returns the state also in case of an error to allow for debugging
// - flow state when graph is requesting prompt or is finished
// - error if any internal error occurred
func Run(flow *model.FlowDefinition, state *model.AuthenticationSession, inputs map[string]string, services *model.Repositories) (*model.AuthenticationSession, error) {

	// Check if state is present and valid
	if state == nil {
		return nil, errors.New("invalid flow state")
	}

	// Check if history size limit is reached
	if len(state.History) > MAX_HISTORY_SIZE {
		return state, errors.New("history size limit reached")
	}

	// If the flow is nil we return an error
	if flow == nil {
		return state, errors.New("invalid flow")
	}

	// If the state is empty we set it to the init node
	if state.Current == "" {
		state.Current = flow.Start
	}

	// Check if node for current state exists in flow
	node, ok := flow.Nodes[state.Current]
	if !ok {
		return state, fmt.Errorf("node '%s' not found in flow", state.Current)
	}

	// Update the current node type
	state.CurrentType = node.Use

	// Load node definition from node name
	def := GetNodeDefinitionByName(node.Use)
	if def == nil {
		return state, fmt.Errorf("node definition for '%s' not found", node.Use)
	}

	var nodeResult *model.NodeResult
	var err error

	// Process node by type
	switch def.Type {
	case model.NodeTypeInit:
		nodeResult, err = ProcessInitTypeNode(state, node, def, inputs, services)

	case model.NodeTypeLogic:
		nodeResult, err = ProcessLogicTypeNode(state, node, def, inputs, services)

	case model.NodeTypeQuery:
		nodeResult, err = ProcessQueryTypeNode(state, node, def, inputs, services)

	case model.NodeTypeResult:
		nodeResult, err = ProcessResultTypeNode(state, node, def, inputs, services)

	case model.NodeTypeQueryWithLogic:
		nodeResult, err = ProcessQueryWithLogicTypeNode(state, node, def, inputs, services)

	default:
		return state, fmt.Errorf("unsupported node type: %s", def.Type)
	}

	userid := ""
	if state.User != nil {
		userid = state.User.ID
	}

	// Return error if present
	if err != nil {

		log.Debug().
			Err(err).
			Str("node_id", node.Name).
			Str("node_type", string(def.Type)).
			Str("user_id", userid).
			Msg("error processing node")
		return state, err
	}

	// End the graph if the node is a result node
	if def.Type == model.NodeTypeResult {

		return state, nil
	}

	// if there are prompt in the result we update the state and return
	if nodeResult.Prompts != nil {

		// turn the nodeResult.Prompts into a strong for logging
		promptsString, err := json.Marshal(nodeResult.Prompts)
		if err != nil {
			log.Debug().Err(err).Msg("error marshalling prompts")
			return nil, err
		}

		// log the node name, type and prompts
		log.Debug().
			Str("node", node.Name).
			Str("node_type", string(def.Type)).
			Str("prompts", string(promptsString)).
			Msg("node resulted in prompts")
		state.History = append(state.History, fmt.Sprintf("%s:prompted:%s", node.Name, promptsString))

		// Update prompts in string and return
		state.Prompts = nodeResult.Prompts
		return state, nil
	}
	if nodeResult.Condition != "" {

		// log the node name and condition
		condition := nodeResult.Condition
		state.History = append(state.History, fmt.Sprintf("%s:%s", node.Name, condition))

		// Check if resulting condition is valid as defined in the node Definition
		valid := false
		for _, c := range def.PossibleResultStates {
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

			log.Debug().Str("node_name", node.Name).Str("condition", condition).Str("flow_id", state.FlowId).Msg("node transition")
		} else {

			log.Debug().Str("node_name", node.Name).Str("condition", condition).Str("flow_id", state.FlowId).Msg("node transition")

			// If we have no next node defined we search for an failureResult node
			// TODO we should log that
			foundFailureResult := false

			// Go through all nodes until we find one of type failureResult
			for nodeName, node := range flow.Nodes {
				if node.Use == "failureResult" {

					// Overwrite the current node with the failureResult node
					state.Error = &[]string{"Invalid node transition"}[0]
					state.Current = nodeName
					state.CurrentType = model.NODE_ERROR
					foundFailureResult = true
					break
				}
			}

			if !foundFailureResult {
				return nil, fmt.Errorf("no next node and no failureResult node defined for condition '%s'", condition)
			}
		}

		// Clear inputs as next node does not expect any inputs
		inputs = nil

		// recursively call run with the new state
		return Run(flow, state, inputs, services)
	}

	// throw an error if no condition or prompts are present
	panic(fmt.Sprintf("node '%s' returned neither prompts nor condition", node.Name))
}

// ProcessQueryTypeNode processes a query node
// and returns the next state and any prompts to be shown to the user
func ProcessQueryTypeNode(state *model.AuthenticationSession, node *model.GraphNode, def *model.NodeDefinition, inputs map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// If no inputs are present send prompts to user
	if inputs == nil {

		return &model.NodeResult{Prompts: def.PossiblePrompts, Condition: ""}, nil
	}

	// Else if we have inputs to context and return submitted
	for k, v := range inputs {
		state.Context[k] = v
	}
	return &model.NodeResult{Prompts: nil, Condition: "submitted"}, nil
}

func ProcessResultTypeNode(state *model.AuthenticationSession, node *model.GraphNode, def *model.NodeDefinition, inputs map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// we expect no inputs for a result node
	if inputs != nil {
		return nil, fmt.Errorf("result node '%s' must not have inputs", node.Name)
	}

	// run the node logic
	// Check if den.Run is not nil
	if def.Run == nil {
		return nil, fmt.Errorf("result node '%s' has no run function", node.Name)
	}

	result, err := def.Run(state, node, nil, services)

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
func ProcessInitTypeNode(state *model.AuthenticationSession, node *model.GraphNode, def *model.NodeDefinition, inputs map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Run init node logic
	result, err := def.Run(state, node, inputs, services)

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
func ProcessLogicTypeNode(state *model.AuthenticationSession, node *model.GraphNode, def *model.NodeDefinition, inputs map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Run node logic
	result, err := def.Run(state, node, inputs, services)
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
func ProcessQueryWithLogicTypeNode(state *model.AuthenticationSession, node *model.GraphNode, def *model.NodeDefinition, inputs map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Run node logic
	result, err := def.Run(state, node, inputs, services)

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
