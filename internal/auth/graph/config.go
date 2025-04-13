package graph

var NodeDefinitions = map[string]*NodeDefinition{
	InitNode.Name:                     InitNode,
	AskUsernameNode.Name:              AskUsernameNode,
	AskPasswordNode.Name:              AskPasswordNode,
	ValidateUsernamePasswordNode.Name: ValidateUsernamePasswordNode,
	SuccessResultNode.Name:            SuccessResultNode,
	FailureResultNode.Name:            FailureResultNode,
	CheckUsernameAvailableNode.Name:   CheckUsernameAvailableNode,
	CreateUserNode.Name:               CreateUserNode,
	SetVariableNode.Name:              SetVariableNode,
	PasskeyRegisterNode.Name:          PasskeyRegisterNode,
	PasskeysVerifyNode.Name:           PasskeysVerifyNode,
	UnlockAccountNode.Name:            UnlockAccountNode,
	PasskeysCheckUserRegistered.Name:  PasskeysCheckUserRegistered,
}

func GetNodeDefinitionByName(name string) *NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
