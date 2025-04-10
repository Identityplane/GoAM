package graph

var UsernamePasswordAuthFlow = &FlowDefinition{
	Name:  "username_password_auth",
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
				"submitted": "askPassword",
			},
		},
		"askPassword": {
			Name: "askPassword",
			Use:  AskPasswordNode.Name,
			Next: map[string]string{
				"submitted": "validateUsernamePassword",
			},
		},
		"validateUsernamePassword": {
			Name: "validateUsernamePassword",
			Use:  ValidateUsernamePasswordNode.Name,
			Next: map[string]string{
				"success": "authSuccess",
				"fail":    "authFailure",
				"locked":  "authFailure",
			},
		},
		"authSuccess": {
			Name: "authSuccess",
			Use:  SuccessResultNode.Name,
			Next: map[string]string{},
			CustomConfig: map[string]string{
				"message": "Login successful!",
			},
		},
		"authFailure": {
			Name: "authFailure",
			Use:  FailureResultNode.Name,
			Next: map[string]string{},
			CustomConfig: map[string]string{
				"message": "Invalid credentials or account locked.",
			},
		},
	},
}
