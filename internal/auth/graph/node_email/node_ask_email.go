package node_email

import "github.com/Identityplane/GoAM/pkg/model"

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
