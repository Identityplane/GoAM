package graph

import (
	"fmt"

	"github.com/Identityplane/GoAM/pkg/model"
)

var InitNode = &model.NodeDefinition{
	Name:                 "init",
	PrettyName:           "Initialize Flow",
	Description:          "Initializes the authentication flow and sets up the starting state",
	Category:             "Flow Control",
	Type:                 model.NodeTypeInit,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"start"},
	Run:                  RunInitNode,
}

var SuccessResultNode = &model.NodeDefinition{
	Name:                 "successResult",
	PrettyName:           "Authentication Success",
	Description:          "Terminal node that indicates successful authentication and returns user information",
	Category:             "Results",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{"user_id"}, // expected to be set by now
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	Run:                  RunAuthSuccessNode,
}

var FailureResultNode = &model.NodeDefinition{
	Name:                 "failureResult",
	PrettyName:           "Authentication Failure",
	Description:          "Terminal node that indicates failed authentication",
	Category:             "Results",
	Type:                 model.NodeTypeResult,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{}, // terminal node
	Run:                  RunAuthFailureNode,
}

func RunAuthFailureNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	state.Result = &model.FlowResult{
		UserID:        "",
		Authenticated: false}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil
}

var MessageConfirmationNode = &model.NodeDefinition{
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
	CustomConfigOptions: map[string]string{
		"message":       "The message to display to the user",
		"message_title": "The title of the message",
		"button_text":   "The text of the button",
	},
}

var AskUsernameNode = &model.NodeDefinition{
	Name:            "askUsername",
	PrettyName:      "Ask for Username",
	Description:     "Prompts the user to enter their username for authentication",
	Category:        "User Input",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"username"},
	PossiblePrompts: map[string]string{
		"username": "text",
	},
	PossibleResultStates: []string{"submitted"},
}

var AskEmailNode = &model.NodeDefinition{
	Name:            "askEmail",
	PrettyName:      "Ask for Email",
	Description:     "Prompts the user to enter their email address",
	Category:        "User Input",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"email"},
	PossiblePrompts: map[string]string{
		"email": "email",
	},
	PossibleResultStates: []string{"submitted"},
}

var AskPasswordNode = &model.NodeDefinition{
	Name:            "askPassword",
	PrettyName:      "Ask for Password",
	Description:     "Prompts the user to enter their password securely",
	Category:        "User Input",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"password"},
	PossiblePrompts: map[string]string{
		"password": "password",
	},
	PossibleResultStates: []string{"submitted"},
}

var SetVariableNode = &model.NodeDefinition{
	Name:                 "setVariable",
	PrettyName:           "Set Variable",
	Description:          "Sets a variable in the flow context with a specified key and value",
	Category:             "Flow Control",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{},
	PossibleResultStates: []string{"done"},
	CustomConfigOptions: map[string]string{
		"key":   "The key to set in the context (required)",
		"value": "The value to set for the key in the context (required)",
	},
	Run: RunSetVariableNode,
}

func RunInitNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	return model.NewNodeResultWithCondition("start")
}

func RunSetVariableNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

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

func RunAuthSuccessNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// If there is a user in the context we use this one
	if state.User != nil {

		state.Result = &model.FlowResult{
			UserID:        state.User.ID,
			Authenticated: true}

		return &model.NodeResult{
			Condition: "",
			Prompts:   nil,
		}, nil
	}

	// Else we check if we have a user_id in the context
	if state.Context["user_id"] != "" {
		state.Result = &model.FlowResult{
			UserID:        state.Context["user_id"],
			Authenticated: true}

		return &model.NodeResult{
			Condition: "",
			Prompts:   nil,
		}, nil
	}

	// TODO we should not reach here as we should have a user or user_id in the context
	// But currently we need it for testing
	state.Result = &model.FlowResult{
		UserID:        "",
		Authenticated: false}

	return &model.NodeResult{
		Condition: "",
		Prompts:   nil,
	}, nil

}

func ptr[T any](v T) *T {
	return &v
}
