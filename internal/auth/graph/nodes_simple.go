package graph

import (
	"goiam/internal/auth/repository"
	"goiam/internal/model"
)

var InitNode = &NodeDefinition{
	Name:            "init",
	Type:            model.NodeTypeInit,
	RequiredContext: []string{},
	OutputContext:   []string{},
	Conditions:      []string{"start"},
	Run:             RunInitNode,
}

var SuccessResultNode = &NodeDefinition{
	Name:            "successResult",
	Type:            model.NodeTypeResult,
	RequiredContext: []string{"user_id", "username"}, // expected to be set by now
	OutputContext:   []string{},
	Conditions:      []string{}, // terminal node
	Run:             RunAuthSuccessNode,
}

func RunAuthSuccessNode(state *model.FlowState, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	state.Result = &model.FlowResult{
		UserID:        state.Context["user_id"],
		Username:      state.Context["username"],
		Authenticated: true}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil

}

var FailureResultNode = &NodeDefinition{
	Name:            "failureResult",
	Type:            model.NodeTypeResult,
	RequiredContext: []string{},
	OutputContext:   []string{},
	Conditions:      []string{}, // terminal node
	Run:             RunAuthFailureNode,
}

func RunAuthFailureNode(state *model.FlowState, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {
	state.Result = &model.FlowResult{
		UserID:        "",
		Username:      "",
		Authenticated: false}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil
}

var AskUsernameNode = &NodeDefinition{
	Name:            "askUsername",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"username"},
	Prompts: map[string]string{
		"username": "text",
	},
	Conditions: []string{"submitted"},
}

var AskPasswordNode = &NodeDefinition{
	Name:            "askPassword",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"password"},
	Prompts: map[string]string{
		"password": "password",
	},
	Conditions: []string{"submitted"},
}

var SetVariableNode = &NodeDefinition{
	Name:            "setVariable",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{},
	OutputContext:   []string{},
	Conditions:      []string{"done"},
	Run:             RunSetVariableNode,
}

func RunInitNode(state *model.FlowState, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	return model.NewNodeResultWithCondition("start")
}

func RunSetVariableNode(state *model.FlowState, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	key := node.CustomConfig["key"]
	value := node.CustomConfig["value"]

	state.Context[key] = value

	return model.NewNodeResultWithCondition("done")
}

func ptr[T any](v T) *T {
	return &v
}
