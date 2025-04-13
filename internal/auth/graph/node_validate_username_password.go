package graph

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ValidateUsernamePasswordNode = &NodeDefinition{
	Name:            "validateUsernamePassword",
	Type:            NodeTypeLogic,
	RequiredContext: []string{"username", "password"},
	OutputContext:   []string{"auth_result"}, // or we may skip outputs if conditions imply it
	Conditions:      []string{"success", "fail", "locked"},
	Run:             RunValidateUsernamePasswordNode,
}

func RunValidateUsernamePasswordNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {
	username := state.Context["username"]
	password := state.Context["password"]

	ctx := context.Background()
	user, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return NewNodeResultWithCondition("fail")
	}

	if user.FailedLoginAttempts >= 3 || user.AccountLocked {
		return NewNodeResultWithCondition("locked")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 3 {
			user.AccountLocked = true
		}
		_ = Services.UserRepo.Update(ctx, user)
		return NewNodeResultWithCondition("fail")
	}

	user.FailedLoginAttempts = 0
	user.AccountLocked = false
	user.LastLoginAt = ptr(time.Now())
	_ = Services.UserRepo.Update(ctx, user)

	state.User = user
	state.Context["auth_result"] = "success"
	state.Context["user_id"] = user.ID

	return NewNodeResultWithCondition("success")
}
