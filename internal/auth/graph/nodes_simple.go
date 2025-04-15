package graph

var InitNode = &NodeDefinition{
	Name:            "init",
	Type:            NodeTypeInit,
	RequiredContext: []string{},
	OutputContext:   []string{},
	Conditions:      []string{"start"},
	Run:             RunInitNode,
}

var SuccessResultNode = &NodeDefinition{
	Name:            "successResult",
	Type:            NodeTypeResult,
	RequiredContext: []string{"user_id", "username"}, // expected to be set by now
	OutputContext:   []string{},
	Conditions:      []string{}, // terminal node
	Run:             RunAuthSuccessNode,
}

func RunAuthSuccessNode(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) {

	state.Result = &FlowResult{
		UserID:        state.Context["user_id"],
		Username:      state.Context["username"],
		Authenticated: true}

	return &NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil

}

var FailureResultNode = &NodeDefinition{
	Name:            "failureResult",
	Type:            NodeTypeResult,
	RequiredContext: []string{},
	OutputContext:   []string{},
	Conditions:      []string{}, // terminal node
	Run:             RunAuthFailureNode,
}

func RunAuthFailureNode(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) {
	state.Result = &FlowResult{
		UserID:        "",
		Username:      "",
		Authenticated: false}

	return &NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil
}

var AskUsernameNode = &NodeDefinition{
	Name:            "askUsername",
	Type:            NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"username"},
	Prompts: map[string]string{
		"username": "text",
	},
	Conditions: []string{"submitted"},
}

var AskPasswordNode = &NodeDefinition{
	Name:            "askPassword",
	Type:            NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"password"},
	Prompts: map[string]string{
		"password": "password",
	},
	Conditions: []string{"submitted"},
}

var SetVariableNode = &NodeDefinition{
	Name:            "setVariable",
	Type:            NodeTypeLogic,
	RequiredContext: []string{},
	OutputContext:   []string{},
	Conditions:      []string{"done"},
	Run:             RunSetVariableNode,
}

func RunInitNode(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) {

	return NewNodeResultWithCondition("start")
}

func RunSetVariableNode(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) {

	key := node.CustomConfig["key"]
	value := node.CustomConfig["value"]

	state.Context[key] = value

	return NewNodeResultWithCondition("done")
}

func ptr[T any](v T) *T {
	return &v
}
