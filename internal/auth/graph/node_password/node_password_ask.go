package node_password

import (
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

var AskPasswordNode = &model.NodeDefinition{
	Name:            "askPassword",
	PrettyName:      "Ask for Password",
	Description:     "Prompts the user to enter their password securely",
	Category:        "User Input",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	OutputContext:   []string{"password"},
	PossiblePrompts: map[string]string{
		"password": "password",
	},
	PossibleResultStates: []string{"submitted"},
	CustomConfigOptions: map[string]string{
		"minLength": "Minimum password length when setting a new password. If unset, only empty passwords are rejected (login can still accept short legacy passwords).",
		"maxLength": "Maximum password length when setting a new password. If unset, no maximum is applied at this prompt (login can still accept long legacy passwords).",
	},
	Run: RunAskPasswordNode,
}

func RunAskPasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	prompts := map[string]string{"password": "password"}
	password, submitted := input["password"]
	if !submitted {
		return model.NewNodeResultWithPrompts(prompts)
	}

	limits := lib.PasswordLimits{}
	if node != nil {
		limits = lib.PasswordLimitsForAsk(node.CustomConfig)
	}
	if err := lib.ValidateNewPassword(password, limits); err != nil {
		state.Error = lib.StringPtr(err.Error())
		return model.NewNodeResultWithPrompts(prompts)
	}

	state.Error = nil
	if state.Context == nil {
		state.Context = map[string]string{}
	}
	state.Context["password"] = password
	return model.NewNodeResultWithCondition("submitted")
}
