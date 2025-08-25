package node_password

import "github.com/Identityplane/GoAM/pkg/model"

var AskUsernamePasswordNode = &model.NodeDefinition{
	Name:            "askUsernamePassword",
	PrettyName:      "Ask for Username and Password",
	Description:     "Prompts the user to enter their username and password securely",
	Category:        "User Input",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{"username", "password"},
	PossiblePrompts: map[string]string{
		"username": "username",
		"password": "password",
	},
	PossibleResultStates: []string{"submitted"},
}
