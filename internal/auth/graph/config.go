package graph

import (
	nodetotp "github.com/Identityplane/GoAM/internal/auth/graph/node_totp"
	"github.com/Identityplane/GoAM/pkg/model"
)

var NodeDefinitions = map[string]*model.NodeDefinition{
	InitNode.Name:                   InitNode,
	AskUsernameNode.Name:            AskUsernameNode,
	AskPasswordNode.Name:            AskPasswordNode,
	SuccessResultNode.Name:          SuccessResultNode,
	FailureResultNode.Name:          FailureResultNode,
	CheckUsernameAvailableNode.Name: CheckUsernameAvailableNode,
	CreateUserNode.Name:             CreateUserNode,
	SetVariableNode.Name:            SetVariableNode,
	UnlockAccountNode.Name:          UnlockAccountNode,
	MessageConfirmationNode.Name:    MessageConfirmationNode,

	// Passkeys
	AskEnrollPasskeyNode.Name:        AskEnrollPasskeyNode,
	PasskeyRegisterNode.Name:         PasskeyRegisterNode,
	PasskeysVerifyNode.Name:          PasskeysVerifyNode,
	PasskeysCheckUserRegistered.Name: PasskeysCheckUserRegistered,
	PasskeyOnboardingNode.Name:       PasskeyOnboardingNode,

	// User Management
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

	// Social Login
	GithubLoginNode.Name:      GithubLoginNode,
	GithubCreateUserNode.Name: GithubCreateUserNode,
	TelegramLoginNode.Name:    TelegramLoginNode,

	// Debug
	DebugNode.Name: DebugNode,
}

func GetNodeDefinitionByName(name string) *model.NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
