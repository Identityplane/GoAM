package graph

import (
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/model"
)

var InitNode = &NodeDefinition{
	Name:                 "init",
	Type:                 model.NodeTypeInit,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"start"},
	Run:                  RunInitNode,
}

var SuccessResultNode = &NodeDefinition{
	Name:                 "successResult",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{"user_id", "username"}, // expected to be set by now
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	Run:                  RunAuthSuccessNode,
}

var FailureResultNode = &NodeDefinition{
	Name:                 "failureResult",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	Run:                  RunAuthFailureNode,
}

func RunAuthFailureNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {
	state.Result = &model.FlowResult{
		UserID:        "",
		Username:      "",
		Authenticated: false}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil
}

var MessageConfirmationNode = &NodeDefinition{
	Name:            "messageConfirmation",
	PrettyName:      "Conformation Dialog",
	Description:     "Display a message to the user and ask for confirmation",
	Category:        "Information",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{},
	PossiblePrompts: map[string]string{
		"confirmation": "boolean",
	},
	PossibleResultStates: []string{"submitted"},
	CustomConfigOptions:  []string{"message", "message_title", "button_text"},
}

var AskUsernameNode = &NodeDefinition{
	Name:            "askUsername",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"username"},
	PossiblePrompts: map[string]string{
		"username": "text",
	},
	PossibleResultStates: []string{"submitted"},
}

var AskEmailNode = &NodeDefinition{
	Name:            "askEmail",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"email"},
	PossiblePrompts: map[string]string{
		"email": "email",
	},
	PossibleResultStates: []string{"submitted"},
}

var AskPasswordNode = &NodeDefinition{
	Name:            "askPassword",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"password"},
	PossiblePrompts: map[string]string{
		"password": "password",
	},
	PossibleResultStates: []string{"submitted"},
}

var SetVariableNode = &NodeDefinition{
	Name:                 "setVariable",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"done"},
	CustomConfigOptions:  []string{"key(required)", "value(required)"},
	Run:                  RunSetVariableNode,
}

func RunInitNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	return model.NewNodeResultWithCondition("start")
}

func RunSetVariableNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	// check if key is set
	if node.CustomConfig["key"] == "" {
		return nil, fmt.Errorf("key is not set")
	}

	// check if value is set
	if node.CustomConfig["value"] == "" {
		return nil, fmt.Errorf("value is not set")
	}

	key := node.CustomConfig["key"]
	value := node.CustomConfig["value"]

	state.Context[key] = value

	return model.NewNodeResultWithCondition("done")
}

func RunAuthSuccessNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	state.Result = &model.FlowResult{
		UserID:        state.Context["user_id"],
		Username:      state.Context["username"],
		Authenticated: true}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil

}

func ptr[T any](v T) *T {
	return &v
}
