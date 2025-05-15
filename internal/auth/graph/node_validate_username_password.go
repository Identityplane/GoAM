package graph

import (
	"context"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ValidateUsernamePasswordNode = &NodeDefinition{
	Name:                 "validateUsernamePassword",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username", "password"},
	OutputContext:        []string{"auth_result"}, // or we may skip outputs if conditions imply it
	PossibleResultStates: []string{"success", "fail", "locked", "noPassword"},
	CustomConfigOptions:  []string{"maxFailedLoginAttempts", "user_lookup_method"},
	Run:                  RunValidateUsernamePasswordNode,
}

func RunValidateUsernamePasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	username := state.Context["username"]
	password := state.Context["password"]
	email := state.Context["email"]
	loginIdentifier := state.Context["loginIdentifier"]

	ctx := context.Background()

	var user *model.User
	var err error

	userLookupMethod := node.CustomConfig["user_lookup_method"]

	switch userLookupMethod {
	case "email":
		user, err = services.UserRepo.GetByEmail(ctx, email)
	case "loginIdentifier":
		user, err = services.UserRepo.GetByLoginIdentifier(ctx, loginIdentifier)
	default:
		user, err = services.UserRepo.GetByUsername(ctx, username)
	}

	// Check if user exists
	if err != nil || user == nil {
		return model.NewNodeResultWithCondition("fail")
	}

	// Check if user is locked or has too many failed login attempts
	if user.FailedLoginAttemptsPassword >= 3 || user.PasswordLocked {
		return model.NewNodeResultWithCondition("locked")
	}

	// If the user has no password set we need to output no password state
	if user.PasswordCredential == "" {
		return model.NewNodeResultWithCondition("noPassword")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordCredential), []byte(password))
	if err != nil {
		user.FailedLoginAttemptsPassword++
		if user.FailedLoginAttemptsPassword >= 3 {
			user.PasswordLocked = true
		}
		_ = services.UserRepo.Update(ctx, user)
		return model.NewNodeResultWithCondition("fail")
	}

	// Reset failed login attempts and unlock user
	user.FailedLoginAttemptsPassword = 0
	user.PasswordLocked = false
	user.LastLoginAt = ptr(time.Now())
	_ = services.UserRepo.Update(ctx, user)

	state.User = user
	state.Context["auth_result"] = "success"
	state.Context["user_id"] = user.ID

	return model.NewNodeResultWithCondition("success")
}
