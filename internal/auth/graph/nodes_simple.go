package graph

var InitNode = &NodeDefinition{
	Name:       "init",
	Type:       NodeTypeInit,
	Inputs:     []string{},
	Outputs:    []string{},
	Conditions: []string{"start"},
}

var SuccessResultNode = &NodeDefinition{
	Name:       "successResult",
	Type:       NodeTypeResult,
	Inputs:     []string{"user_id", "username"}, // expected to be set by now
	Outputs:    []string{},
	Conditions: []string{}, // terminal node
}

var FailureResultNode = &NodeDefinition{
	Name:       "failureResult",
	Type:       NodeTypeResult,
	Inputs:     []string{},
	Outputs:    []string{},
	Conditions: []string{}, // terminal node
}

var AskUsernameNode = &NodeDefinition{
	Name:    "askUsername",
	Type:    NodeTypeQuery,
	Inputs:  []string{},
	Outputs: []string{"username"},
	Prompts: map[string]string{
		"username": "text",
	},
	Conditions: []string{"submitted"},
}

var AskPasswordNode = &NodeDefinition{
	Name:    "askPassword",
	Type:    NodeTypeQuery,
	Inputs:  []string{},
	Outputs: []string{"password"},
	Prompts: map[string]string{
		"password": "password",
	},
	Conditions: []string{"submitted"},
}

func RunInitNode(state *FlowState, node *GraphNode) (string, error) {

	return "start", nil
}

func ptr[T any](v T) *T {
	return &v
}
