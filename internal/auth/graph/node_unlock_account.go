package graph

import (
	"context"
	"time"
)

// Unlock account logic node
var UnlockAccountNode = &NodeDefinition{
	Name:            "unlockAccount",
	Type:            NodeTypeLogic,
	RequiredContext: []string{"username"},
	OutputContext:   []string{},
	Conditions:      []string{"success", "fail"},
	Run:             RunUnlockAccountNode,
}

func RunUnlockAccountNode(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) {
	username := state.Context["username"]

	ctx := context.Background()
	user, err := services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return NewNodeResultWithCondition("fail")
	}

	if !user.PasswordLocked {
		return NewNodeResultWithCondition("success")
	}

	user.PasswordLocked = false
	user.FailedLoginAttemptsPassword = 0
	user.UpdatedAt = time.Now()

	if err := services.UserRepo.Update(ctx, user); err != nil {
		return NewNodeResultWithCondition("fail")
	}

	return NewNodeResultWithCondition("success")
}
