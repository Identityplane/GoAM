package graph

var NodeDefinitions = map[string]*NodeDefinition{
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

	// User Management
	LoadUserByUsernameNode.Name: LoadUserByUsernameNode,

	// Password or Social Login
	PasswordOrSocialLoginNode.Name: PasswordOrSocialLoginNode,

	// Email OTP
	AskEmailNode.Name: AskEmailNode,
	EmailOTPNode.Name: EmailOTPNode,

	// Password
	UpdatePasswordNode.Name:           UpdatePasswordNode,
	ValidateUsernamePasswordNode.Name: ValidateUsernamePasswordNode,

	// Hcaptcha
	HcaptchaNode.Name: HcaptchaNode,

	// Social Login
	GithubLoginNode.Name:      GithubLoginNode,
	GithubCreateUserNode.Name: GithubCreateUserNode,
}

func GetNodeDefinitionByName(name string) *NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
