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

func RunValidateUsernamePasswordNode(state *FlowState, node *GraphNode, input map[string]string, services *ServiceRegistry) (*NodeResult, error) {
	username := state.Context["username"]
	password := state.Context["password"]

	ctx := context.Background()
	user, err := services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return NewNodeResultWithCondition("fail")
	}

	if user.FailedLoginAttemptsPassword >= 3 || user.PasswordLocked {
		return NewNodeResultWithCondition("locked")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordCredential), []byte(password))
	if err != nil {
		user.FailedLoginAttemptsPassword++
		if user.FailedLoginAttemptsPassword >= 3 {
			user.PasswordLocked = true
		}
		_ = services.UserRepo.Update(ctx, user)
		return NewNodeResultWithCondition("fail")
	}

	user.FailedLoginAttemptsPassword = 0
	user.PasswordLocked = false
	user.LastLoginAt = ptr(time.Now())
	_ = services.UserRepo.Update(ctx, user)

	state.User = user
	state.Context["auth_result"] = "success"
	state.Context["user_id"] = user.ID

	return NewNodeResultWithCondition("success")
}
