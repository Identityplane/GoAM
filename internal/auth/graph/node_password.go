package graph

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/model"
)

var UpdatePasswordNode = &NodeDefinition{
	Name:                 "updatePassword",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"user"},
	OutputContext:        []string{}, // or we may skip outputs if conditions imply it
	PossibleResultStates: []string{"success", "fail"},
	Run:                  RunUpdatePasswordNode,
}

var ValidateUsernamePasswordNode = &NodeDefinition{
	Name:                 "validateUsernamePassword",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username", "password"},
	OutputContext:        []string{"auth_result"}, // or we may skip outputs if conditions imply it
	PossibleResultStates: []string{"success", "fail", "locked", "noPassword"},
	CustomConfigOptions: map[string]string{
		"max_failed_password_attempts": "Maximum number of failed password attempts before locking the user (default: 10)",
		"user_lookup_method":           "Method to look up user: 'username', 'email', or 'loginIdentifier' (default: username)",
	},
	Run: RunValidateUsernamePasswordNode,
}

func RunValidateUsernamePasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	username := state.Context["username"]
	password := state.Context["password"]
	email := state.Context["email"]
	loginIdentifier := state.Context["loginIdentifier"]

	// if max_failed_password_attempts is set use else default to 5
	maxFailedPasswordAttempts := 10
	var err error
	if node.CustomConfig["max_failed_password_attempts"] != "" {
		maxFailedPasswordAttempts, err = strconv.Atoi(node.CustomConfig["max_failed_password_attempts"])
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to convert max_failed_password_attempts to int: %w", err))
		}
	}

	ctx := context.Background()

	var user *model.User

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
	if user.FailedLoginAttemptsPassword >= maxFailedPasswordAttempts || user.PasswordLocked {
		return model.NewNodeResultWithCondition("locked")
	}

	// If the user has no password set we need to output no password state
	if user.PasswordCredential == "" {
		return model.NewNodeResultWithCondition("noPassword")
	}

	// Compare password
	err = lib.ComparePassword(password, user.PasswordCredential)
	if err != nil {
		user.FailedLoginAttemptsPassword++
		if user.FailedLoginAttemptsPassword >= maxFailedPasswordAttempts {
			user.PasswordLocked = true
		}
		_ = services.UserRepo.Update(ctx, user)

		if user.FailedLoginAttemptsPassword >= maxFailedPasswordAttempts {
			return model.NewNodeResultWithCondition("locked")
		}

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

func RunUpdatePasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	if state.User == nil {
		return model.NewNodeResultWithError(errors.New("user must be loaded before updating password"))
	}

	// If the password is not set in the input, use the password from the context
	if input["password"] != "" {
		state.Context["password"] = input["password"]
	}

	if state.Context["password"] == "" {

		return model.NewNodeResultWithPrompts(map[string]string{"password": "password"})
	}

	hashed, err := lib.HashPassword(state.Context["password"])
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to hash password: %w", err))
	}

	state.User.PasswordCredential = hashed
	err = services.UserRepo.Update(context.Background(), state.User)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to update password: %w", err))
	}

	return model.NewNodeResultWithCondition("success")
}
