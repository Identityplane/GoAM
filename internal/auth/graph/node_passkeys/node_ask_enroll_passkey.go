package node_passkeys

import "github.com/Identityplane/GoAM/pkg/model"

var AskEnrollPasskeyNode = &model.NodeDefinition{
	Name:                 "askEnrollPasskey",
	PrettyName:           "Ask to Enroll Passkey",
	Description:          "Prompts the user to choose whether they want to enroll a passkey for future passwordless authentication",
	Category:             "Passkeys",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{},
	PossiblePrompts:      map[string]string{"enrollPasskey": "boolean"},
	OutputContext:        []string{"enrollPasskey"},
	PossibleResultStates: []string{"yes", "no"},
	Run:                  RunAskEnrollPasskeyNode,
}

// Very simple node that asks the user if they want to enroll a passkey
func RunAskEnrollPasskeyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	enrollPasskey := input["enrollPasskey"]

	if enrollPasskey == "" {
		return model.NewNodeResultWithPrompts(map[string]string{"enrollPasskey": "boolean"})
	} else if enrollPasskey == "true" {
		return model.NewNodeResultWithCondition("yes")
	} else {
		return model.NewNodeResultWithCondition("no")
	}

}
