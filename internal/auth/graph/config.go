package graph

import (
	"github.com/Identityplane/GoAM/internal/auth/graph/node_github"
	nodetotp "github.com/Identityplane/GoAM/internal/auth/graph/node_totp"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_user"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_username"
	"github.com/Identityplane/GoAM/pkg/model"
)

var NodeDefinitions = map[string]*model.NodeDefinition{
	InitNode.Name:                InitNode,
	AskUsernameNode.Name:         AskUsernameNode,
	AskPasswordNode.Name:         AskPasswordNode,
	SuccessResultNode.Name:       SuccessResultNode,
	FailureResultNode.Name:       FailureResultNode,
	CreateUserNode.Name:          CreateUserNode,
	SetVariableNode.Name:         SetVariableNode,
	UnlockAccountNode.Name:       UnlockAccountNode,
	MessageConfirmationNode.Name: MessageConfirmationNode,

	// Username
	node_username.CheckUsernameAvailableNode.Name: node_username.CheckUsernameAvailableNode,

	// Passkeys
	AskEnrollPasskeyNode.Name:        AskEnrollPasskeyNode,
	PasskeyRegisterNode.Name:         PasskeyRegisterNode,
	PasskeysVerifyNode.Name:          PasskeysVerifyNode,
	PasskeysCheckUserRegistered.Name: PasskeysCheckUserRegistered,
	PasskeyOnboardingNode.Name:       PasskeyOnboardingNode,

	// User Management
	node_user.InitUserNode.Name: node_user.InitUserNode,
	LoadUserByUsernameNode.Name: LoadUserByUsernameNode,

	// Password or Social Login
	PasswordOrSocialLoginNode.Name: PasswordOrSocialLoginNode,

	// MFA
	AskEmailNode.Name:            AskEmailNode,
	EmailOTPNode.Name:            EmailOTPNode,
	nodetotp.TOTPCreateNode.Name: nodetotp.TOTPCreateNode,
	nodetotp.TOTPVerifyNode.Name: nodetotp.TOTPVerifyNode,

	// Password
	UpdatePasswordNode.Name:           UpdatePasswordNode,
	ValidateUsernamePasswordNode.Name: ValidateUsernamePasswordNode,

	// Hcaptcha
	HcaptchaNode.Name: HcaptchaNode,

	// Social Logins
	TelegramLoginNode.Name: TelegramLoginNode,

	// GitHub
	node_github.GithubLoginNode.Name:      node_github.GithubLoginNode,
	node_github.GithubCreateUserNode.Name: node_github.GithubCreateUserNode,

	// Debug
	DebugNode.Name: DebugNode,
}

func GetNodeDefinitionByName(name string) *model.NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
