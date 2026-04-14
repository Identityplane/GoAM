package node_qr

import (
	"github.com/Identityplane/GoAM/pkg/model"
)

const continueOnOtherDevicePromptMessage = "message"

var ContinueOnOtherDeviceNode = &model.NodeDefinition{
	Name:          "continueOnOtherDevice",
	PrettyName:    "Continue on other device",
	Description:   "Updates shared context so the primary device can proceed, then shows a success message on this device; acknowledgement moves to successResult.",
	Category:      "QR",
	Type:          model.NodeTypeQueryWithLogic,
	OutputContext: []string{"message"},
	PossibleResultStates: []string{
		"done",
	},
	Run: RunContinueOnOtherDeviceNode,
}

func RunContinueOnOtherDeviceNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	if state.Context == nil {
		state.Context = make(map[string]string)
	}

	// User dismissed the message — advance to successResult (context was already updated on first visit).
	if input != nil && input["ack"] == "true" {
		return model.NewNodeResultWithCondition("done")
	}

	// First: update primary so polling can complete; then show the success-style message in the same step.
	state.Context[CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE] = "true"
	return model.NewNodeResultWithPrompts(map[string]string{
		continueOnOtherDevicePromptMessage: "Success — you can continue signing in on your computer. Your browser will finish signing in. Tap continue when you are ready to close this screen.",
	})
}
