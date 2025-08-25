package node_username

import "github.com/Identityplane/GoAM/pkg/model"

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
