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

func RunUnlockAccountNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {
	username := state.Context["username"]

	ctx := context.Background()
	user, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return NewNodeResultWithCondition("fail")
	}

	if !user.AccountLocked {
		return NewNodeResultWithCondition("success")
	}

	user.AccountLocked = false
	user.FailedLoginAttempts = 0
	user.UpdatedAt = time.Now()

	if err := Services.UserRepo.Update(ctx, user); err != nil {
		return NewNodeResultWithCondition("fail")
	}

	return NewNodeResultWithCondition("success")
}
