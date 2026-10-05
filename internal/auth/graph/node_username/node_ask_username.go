package node_username

import (
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

var AskUsernameNode = &model.NodeDefinition{
	Name:            "askUsername",
	PrettyName:      "Ask for Username",
	Description:     "Prompts the user to enter their username for authentication",
	Category:        "User Input",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	OutputContext:   []string{"username"},
	PossiblePrompts: map[string]string{
		"username": "text",
	},
	PossibleResultStates: []string{"submitted"},
	Run:                  RunAskUsernameNode,
}

func RunAskUsernameNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	prompts := map[string]string{"username": "text"}
	username, submitted := input["username"]
	if !submitted {
		return model.NewNodeResultWithPrompts(prompts)
	}

	if err := lib.ValidateUsername(username); err != nil {
		state.Error = lib.StringPtr(err.Error())
		return model.NewNodeResultWithPrompts(prompts)
	}

	if state.Context == nil {
		state.Context = map[string]string{}
	}
	state.Error = nil
	state.Context["username"] = lib.NormalizeUsername(username)
	return model.NewNodeResultWithCondition("submitted")
}
