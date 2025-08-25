package node_password

import "github.com/Identityplane/GoAM/pkg/model"

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
