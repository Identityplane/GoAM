package graph

import (
	"fmt"

	"github.com/gianlucafrei/GoAM/internal/auth/repository"
	"github.com/gianlucafrei/GoAM/internal/model"
)

var PasswordOrSocialLoginNode = &NodeDefinition{
	Name:                 "passwordOrSocialLogin",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{""},
	PossiblePrompts:      map[string]string{"option": "text", "username": "text", "password": "password", "email": "email", "passkeysLoginOptions": "json", "passkeysFinishLoginJson": "json"},
	OutputContext:        []string{"username", "password"},
	PossibleResultStates: []string{"password", "forgotPassword", "passkey", "social1", "social2", "social3"},
	CustomConfigOptions:  []string{"useEmail", "showForgotPassword", "showPasskeys", "showSocial1", "showSocial2", "social1Provider", "social2Provider"},
	Run:                  RunPasswordOrSocialLoginNode,
}

func RunPasswordOrSocialLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	// if option is not set we return the prompt
	option, ok := input["option"]
	if !ok {

		// For passkey discovery we create a passkey challenge
		passkeysLoginOptions, err := generatePasskeysChallenge(state, "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to generate passkey challenge: %w", err)
		}

		return model.NewNodeResultWithPrompts(map[string]string{"option": "text", "username": "text", "password": "password", "email": "email", "passkeysLoginOptions": passkeysLoginOptions})
	}

	// if option is set we return the result
	switch option {
	case "password":

		// Copy username and password to context
		state.Context["username"] = input["username"]
		state.Context["email"] = input["email"]
		state.Context["password"] = input["password"]

		return model.NewNodeResultWithCondition("password")
	case "forgot-password":
		return model.NewNodeResultWithCondition("forgotPassword")
	case "passkey":

		// If we have a passkeyFinishLoginJson we add it to the context
		// This is the case if the user uses passkey discovery login
		if input["passkeysFinishLoginJson"] != "" {
			state.Context["passkeysFinishLoginJson"] = input["passkeysFinishLoginJson"]
		}

		return model.NewNodeResultWithCondition("passkey")
	case "social1":
		return model.NewNodeResultWithCondition("social1")
	case "social2":
		return model.NewNodeResultWithCondition("social2")
	}

	return nil, fmt.Errorf("invalid option: %s", option)
}
