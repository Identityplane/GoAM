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
}

var LogicFunctions = map[string]LogicFunc{
	InitNode.Name:                     RunInitNode,
	ValidateUsernamePasswordNode.Name: RunValidateUsernamePasswordNode,
	CreateUserNode.Name:               RunCreateUserNode,
	CheckUsernameAvailableNode.Name:   RunCheckUsernameAvailableNode,
}

func GetNodeDefinitionByName(name string) *NodeDefinition {
	if def, ok := NodeDefinitions[name]; ok {
		return def
	}
	return nil
}
