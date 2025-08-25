package node_password

import "github.com/Identityplane/GoAM/pkg/model"

var AskEmailPasswordNode = &model.NodeDefinition{
	Name:            "askEmailPassword",
	PrettyName:      "Ask for Email and Password",
	Description:     "Prompts the user to enter their email and password securely",
	Category:        "User Input",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"email", "password"},
	PossiblePrompts: map[string]string{
		"email":    "email",
		"password": "password",
	},
	PossibleResultStates: []string{"submitted"},
}
