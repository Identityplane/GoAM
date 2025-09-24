package node_options

import (
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_passkeys"
	"github.com/Identityplane/GoAM/pkg/model"
)

var PasswordOrSocialLoginNode = &model.NodeDefinition{
	Name:                 "passwordOrSocialLogin",
	PrettyName:           "Password or Social Login",
	Description:          "This node is used to login with password or social login",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{""},
	PossiblePrompts:      map[string]string{"option": "text", "username": "text", "password": "password", "email": "email", "passkeysLoginOptions": "json", "passkeysFinishLoginJson": "json"},
	OutputContext:        []string{"username", "password"},
	PossibleResultStates: []string{"password", "forgotPassword", "passkey", "social1", "social2", "social3", "register"},
	CustomConfigOptions: map[string]string{
		"showRegistrationLink": "if 'true' then show registration link, otherwise hide it",
		"useUsername":          "if 'true' then show username input, otherwise hide it",
		"useEmail":             "if 'true' then show email input, otherwise will as for username as input",
		"usePassword":          "if 'true' then show password input, otherwise hide it",
		"usePasskeys":          "if 'true' then show passkeys input (including discovery), otherwise hide it",
		"showForgotPassword":   "if 'true' then show forgot password input, otherwise hide it",
		"disableSubmit":        "if 'true' then hide submit button, otherwise show it",
		"disablePasskeyBtn":    "if 'true' then hide passkey button, otherwise show it",
		"social1":              "If set shows the login button for the social login provider 1 with the specified provider name.",
		"social2":              "If set shows the login button for the social login provider 2 with the specified provider name.",
		"social3":              "If set shows the login button for the social login provider 3 with the specified provider name.",
		"title":                "Title of the login form",
		"message":              "Message of the login form",
	},
	Run: RunPasswordOrSocialLoginNode,
}

func RunPasswordOrSocialLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// if option is not set we return the prompt
	// check if starts with passwordOrSocialLogin:prompted
	latestHistory := state.GetLatestHistory()
	latestIsOptionsNdoe := strings.HasPrefix(latestHistory, "passwordOrSocialLogin:prompted")

	option, ok := input["option"]
	if !latestIsOptionsNdoe || !ok {

		prompts := map[string]string{
			"option": "text",
		}

		if node.CustomConfig["usePasskeys"] == "true" {
			// For passkey discovery we create a passkey challenge
			passkeysLoginOptions, err := node_passkeys.GeneratePasskeysChallenge(state, node, "", "")
			if err != nil {
				return nil, fmt.Errorf("failed to generate passkey challenge: %w", err)
			}
			prompts["passkeysLoginOptions"] = passkeysLoginOptions
		}

		if node.CustomConfig["useUsername"] == "true" {
			prompts["username"] = "text"
		}
		if node.CustomConfig["useEmail"] == "true" {
			prompts["email"] = "email"
		}
		if node.CustomConfig["usePassword"] == "true" {
			prompts["password"] = "password"
		}
		if node.CustomConfig["showForgotPassword"] == "true" {
			prompts["forgotPassword"] = "forgotPassword"
		}
		if node.CustomConfig["showRegistrationLink"] == "true" {
			prompts["register"] = "register"
		}
		if node.CustomConfig["social1"] != "" {
			prompts["social1"] = node.CustomConfig["social1"]
		}
		if node.CustomConfig["social2"] != "" {
			prompts["social2"] = node.CustomConfig["social2"]
		}
		if node.CustomConfig["social3"] != "" {
			prompts["social3"] = node.CustomConfig["social3"]
		}

		return model.NewNodeResultWithPrompts(prompts)
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
	case "register":
		return model.NewNodeResultWithCondition("register")
	case "social1":
		return model.NewNodeResultWithCondition("social1")
	case "social2":
		return model.NewNodeResultWithCondition("social2")
	}

	return nil, fmt.Errorf("invalid option: %s", option)
}
