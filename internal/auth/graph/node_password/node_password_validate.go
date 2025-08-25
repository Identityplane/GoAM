package node_password

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

var ValidateUsernamePasswordNode = &model.NodeDefinition{
	Name:                 "validateUsernamePassword",
	PrettyName:           "Validate Username and Password",
	Description:          "Validates the provided username and password against the database and handles account locking",
	Category:             "Authentication",
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

func RunValidateUsernamePasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Get password from context
	password := state.Context["password"]
	if password == "" {
		return model.NewNodeResultWithError(fmt.Errorf("password is required in context"))
	}

	// if max_failed_password_attempts is set use else default to 10
	maxFailedPasswordAttempts := 10
	var err error
	if node.CustomConfig["max_failed_password_attempts"] != "" {
		maxFailedPasswordAttempts, err = strconv.Atoi(node.CustomConfig["max_failed_password_attempts"])
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to convert max_failed_password_attempts to int: %w", err))
		}
	}

	ctx := context.Background()
	user, err := node_utils.LoadUserFromContext(state, services)

	// Check if user exists
	if err != nil || user == nil {
		state.Error = stringPtr("Invalid password")
		return model.NewNodeResultWithCondition("fail")
	}

	// Get password attribute
	passwordValue, attribute, err := model.GetAttribute[model.PasswordAttributeValue](user, model.AttributeTypePassword)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to get password attribute: %w", err))
	}

	// If the user has no password set we need to output no password state
	if passwordValue == nil || passwordValue.PasswordHash == "" {
		state.Error = stringPtr("User has no password")
		return model.NewNodeResultWithCondition("noPassword")
	}

	// Check if user is locked or has too many failed login attempts
	if passwordValue.FailedAttempts >= maxFailedPasswordAttempts || passwordValue.Locked {
		state.Error = stringPtr("User is locked")
		return model.NewNodeResultWithCondition("locked")
	}

	// Compare password
	err = lib.ComparePassword(password, passwordValue.PasswordHash)
	if err != nil {

		// Increment failed attempts
		passwordValue.FailedAttempts++
		if passwordValue.FailedAttempts >= maxFailedPasswordAttempts {
			passwordValue.Locked = true
		}

		// Update the password attribute in the user
		for i, attr := range user.UserAttributes {
			if attr.ID == attribute.ID {
				user.UserAttributes[i].Value = passwordValue
				// Update the attribute reference to point to the updated one
				attribute = &user.UserAttributes[i]
				break
			}
		}

		err = services.UserRepo.UpdateUserAttribute(ctx, attribute)
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to update password attribute: %w", err))
		}

		if passwordValue.FailedAttempts >= maxFailedPasswordAttempts || passwordValue.Locked {
			state.Error = stringPtr("User is locked")
			return model.NewNodeResultWithCondition("locked")
		}

		state.Error = stringPtr("Invalid password")
		return model.NewNodeResultWithCondition("fail")
	}

	// Reset failed login attempts and unlock user
	passwordValue.FailedAttempts = 0
	passwordValue.Locked = false
	passwordValue.LastCorrectTimestamp = timePtr(time.Now())

	// Update the password attribute in the user
	for i, attr := range user.UserAttributes {
		if attr.ID == attribute.ID {
			user.UserAttributes[i].Value = passwordValue
			// Update the attribute reference to point to the updated one
			attribute = &user.UserAttributes[i]
			break
		}
	}

	// Update last login time
	user.LastLoginAt = timePtr(time.Now())

	err = services.UserRepo.UpdateUserAttribute(ctx, attribute)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to update password attribute: %w", err))
	}

	state.User = user
	state.Context["auth_result"] = "success"

	return model.NewNodeResultWithCondition("success")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
