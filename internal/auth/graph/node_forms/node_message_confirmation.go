package node_forms

import "github.com/Identityplane/GoAM/pkg/model"

var MessageConfirmationNode = &model.NodeDefinition{
	Name:            "messageConfirmation",
	PrettyName:      "Conformation Dialog",
	Description:     "Display a message to the user and ask for confirmation",
	Category:        "Information",
	Type:            model.NodeTypeQuery,
	RequiredContext: []string{},
	OutputContext:   []string{},
	PossiblePrompts: map[string]string{
		"confirmation": "boolean",
	},
	PossibleResultStates: []string{"submitted"},
	CustomConfigOptions: map[string]string{
		"message":       "The message to display to the user",
		"message_title": "The title of the message",
		"button_text":   "The text of the button",
	},
}
