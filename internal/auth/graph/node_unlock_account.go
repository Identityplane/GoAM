package graph

import (
	"context"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
	"time"
)

// Unlock account logic node
var UnlockAccountNode = &NodeDefinition{
	Name:                 "unlockAccount",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	OutputContext:        []string{},
	PossibleResultStates: []string{"success", "fail"},
	Run:                  RunUnlockAccountNode,
}

func RunUnlockAccountNode(state *model.FlowState, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {
	username := state.Context["username"]

	ctx := context.Background()
	user, err := services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return model.NewNodeResultWithCondition("fail")
	}

	if !user.PasswordLocked {
		return model.NewNodeResultWithCondition("success")
	}

	user.PasswordLocked = false
	user.FailedLoginAttemptsPassword = 0
	user.UpdatedAt = time.Now()

	if err := services.UserRepo.Update(ctx, user); err != nil {
		return model.NewNodeResultWithCondition("fail")
	}

	return model.NewNodeResultWithCondition("success")
}
