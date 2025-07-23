package graph

import (
	"context"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"
)

// Unlock account logic node
var UnlockAccountNode = &model.NodeDefinition{
	Name:                 "unlockAccount",
	PrettyName:           "Unlock Account",
	Description:          "Unlocks a user account that has been locked due to too many failed login attempts",
	Category:             "User Management",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	OutputContext:        []string{},
	PossibleResultStates: []string{"success", "fail"},
	Run:                  RunUnlockAccountNode,
}

func RunUnlockAccountNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
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
