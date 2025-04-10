package graph

var UserRegisterFlow = &FlowDefinition{
	Name:  "user_register",
	Start: "init",
	Nodes: map[string]*GraphNode{
		"init": {
			Name: "init",
			Use:  InitNode.Name,
			Next: map[string]string{
				"start": "askUsername",
			},
		},
		"askUsername": {
			Name: "askUsername",
			Use:  AskUsernameNode.Name,
			Next: map[string]string{
				"submitted": "checkUsernameAvailable",
			},
		},
		"checkUsernameAvailable": {
			Name: "checkUsernameAvailable",
			Use:  CheckUsernameAvailableNode.Name,
			Next: map[string]string{
				"available": "askPassword",
				"taken":     "registerFailed",
			},
		},
		"askPassword": {
			Name: "askPassword",
			Use:  AskPasswordNode.Name,
			Next: map[string]string{
				"submitted": "createUser",
			},
		},
		"createUser": {
			Name: "createUser",
			Use:  CreateUserNode.Name,
			Next: map[string]string{
				"success": "registerSuccess",
				"fail":    "registerFailed",
			},
		},
		"registerSuccess": {
			Name: "registerSuccess",
			Use:  SuccessResultNode.Name,
			CustomConfig: map[string]string{
				"message": "Registration successful!",
			},
		},
		"registerFailed": {
			Name: "registerFailed",
			Use:  FailureResultNode.Name,
			CustomConfig: map[string]string{
				"message": "Registration failed. Username may already exist.",
			},
		},
	},
}
