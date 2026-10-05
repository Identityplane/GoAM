package node_password

import (
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

var AskUsernamePasswordNode = &model.NodeDefinition{
	Name:            "askUsernamePassword",
	PrettyName:      "Ask for Username and Password",
	Description:     "Prompts the user to enter their username and password securely",
	Category:        "User Input",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	OutputContext:   []string{"username", "password"},
	PossiblePrompts: map[string]string{
		"username": "username",
		"password": "password",
	},
	PossibleResultStates: []string{"submitted"},
	Run:                  RunAskUsernamePasswordNode,
}

func RunAskUsernamePasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	prompts := map[string]string{
		"username": "username",
		"password": "password",
	}
	_, hasUsername := input["username"]
	_, hasPassword := input["password"]
	if !hasUsername && !hasPassword {
		return model.NewNodeResultWithPrompts(prompts)
	}

	username := input["username"]
	password := input["password"]

	if err := lib.ValidateUsername(username); err != nil {
		state.Error = lib.StringPtr(err.Error())
		return model.NewNodeResultWithPrompts(prompts)
	}
	if err := lib.ValidateNewPassword(password, lib.PasswordLimits{}); err != nil {
		state.Error = lib.StringPtr(err.Error())
		return model.NewNodeResultWithPrompts(prompts)
	}

	state.Error = nil
	if state.Context == nil {
		state.Context = map[string]string{}
	}
	state.Context["username"] = lib.NormalizeUsername(username)
	state.Context["password"] = password
	return model.NewNodeResultWithCondition("submitted")
}
