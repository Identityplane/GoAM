package graph

import (
	"github.com/Identityplane/GoAM/internal/auth/graph/node_captcha"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_email"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_forms"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_github"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_options"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_passkeys"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_password"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_system"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_telegram"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_totp"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_user"
	"github.com/Identityplane/GoAM/internal/auth/graph/node_username"
	"github.com/Identityplane/GoAM/pkg/model"
)

var NodeDefinitions = map[string]*model.NodeDefinition{

	// System
	node_system.InitNode.Name:          node_system.InitNode,
	node_system.SuccessResultNode.Name: node_system.SuccessResultNode,
	node_system.FailureResultNode.Name: node_system.FailureResultNode,
	node_system.SetVariableNode.Name:   node_system.SetVariableNode,
	node_system.DebugNode.Name:         node_system.DebugNode,

	// User Management
	node_user.CreateUserNode.Name: node_user.CreateUserNode,
	node_user.InitUserNode.Name:   node_user.InitUserNode,
	node_user.LoadUserNode.Name:   node_user.LoadUserNode,
	node_user.SaveUserNode.Name:   node_user.SaveUserNode,

	// Username
	node_username.AskUsernameNode.Name:            node_username.AskUsernameNode,
	node_username.CheckUsernameAvailableNode.Name: node_username.CheckUsernameAvailableNode,

	// Password
	node_password.AskPasswordNode.Name:              node_password.AskPasswordNode,
	node_password.UpdatePasswordNode.Name:           node_password.UpdatePasswordNode,
	node_password.ValidateUsernamePasswordNode.Name: node_password.ValidateUsernamePasswordNode,
	node_password.AskUsernamePasswordNode.Name:      node_password.AskUsernamePasswordNode,
	node_password.AskEmailPasswordNode.Name:         node_password.AskEmailPasswordNode,

	// Forms
	node_forms.MessageConfirmationNode.Name: node_forms.MessageConfirmationNode,

	// Passkeys
	node_passkeys.AskEnrollPasskeyNode.Name:        node_passkeys.AskEnrollPasskeyNode,
	node_passkeys.PasskeyRegisterNode.Name:         node_passkeys.PasskeyRegisterNode,
	node_passkeys.PasskeysVerifyNode.Name:          node_passkeys.PasskeysVerifyNode,
	node_passkeys.PasskeysCheckUserRegistered.Name: node_passkeys.PasskeysCheckUserRegistered,
	node_passkeys.PasskeyOnboardingNode.Name:       node_passkeys.PasskeyOnboardingNode,

	// Password or Social Login
	node_options.PasswordOrSocialLoginNode.Name: node_options.PasswordOrSocialLoginNode,

	// Email
	node_email.AskEmailNode.Name:            node_email.AskEmailNode,
	node_email.CheckEmailAvailableNode.Name: node_email.CheckEmailAvailableNode,
	node_email.EmailOTPNode.Name:            node_email.EmailOTPNode,
	node_email.HasEmailNode.Name:            node_email.HasEmailNode,
	node_email.SaveEmailNode.Name:           node_email.SaveEmailNode,

	// TOTP
	node_totp.TOTPCreateNode.Name: node_totp.TOTPCreateNode,
	node_totp.TOTPVerifyNode.Name: node_totp.TOTPVerifyNode,

	// Captcha
	node_captcha.HcaptchaNode.Name: node_captcha.HcaptchaNode,

	// Social Logins
	node_telegram.TelegramLoginNode.Name: node_telegram.TelegramLoginNode,

	// GitHub
	node_github.GithubLoginNode.Name:      node_github.GithubLoginNode,
	node_github.GithubCreateUserNode.Name: node_github.GithubCreateUserNode,
}

func GetNodeDefinitionByName(name string) *model.NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
