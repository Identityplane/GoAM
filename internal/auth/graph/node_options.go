package graph

import (
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/model"
)

var PasswordOrSocialLoginNode = &NodeDefinition{
	Name:                 "passwordOrSocialLogin",
	PrettyName:           "Password or Social Login",
	Description:          "This node is used to login with password or social login",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{""},
	PossiblePrompts:      map[string]string{"option": "text", "username": "text", "password": "password", "email": "email", "passkeysLoginOptions": "json", "passkeysFinishLoginJson": "json"},
	OutputContext:        []string{"username", "password"},
	PossibleResultStates: []string{"password", "forgotPassword", "passkey", "social1", "social2", "social3"},
	CustomConfigOptions: map[string]string{
		"showRegistrationLink": "if 'true' then show registration link, otherwise hide it",
		"useEmail":             "if 'true' then show email input, otherwise will as for username as input",
		"showForgotPassword":   "if 'true' then show forgot password input, otherwise hide it",
		"showPasskeys":         "if 'true' then show passkeys options, otherwise hide it",
		"social1":              "if 'true' then show social1 input, otherwise hide it",
		"social2":              "if 'true' then show social2 input, otherwise hide it",
		"social1Provider":      "Currently build in are 'google' and 'github'. If not set the default button will be shown",
		"social2Provider":      "Currently build in are 'google' and 'github'. If not set the default button will be shown",
	},
	Run: RunPasswordOrSocialLoginNode,
}

func RunPasswordOrSocialLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	// if option is not set we return the prompt
	// check if starts with passwordOrSocialLogin:prompted
	latestHistory := state.GetLatestHistory()
	latestIsOptionsNdoe := strings.HasPrefix(latestHistory, "passwordOrSocialLogin:prompted")

	option, ok := input["option"]
	if !latestIsOptionsNdoe || !ok {

		// For passkey discovery we create a passkey challenge
		passkeysLoginOptions, err := generatePasskeysChallenge(state, node, "", "")
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
