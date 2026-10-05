package node_password

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

var UpdatePasswordNode = &model.NodeDefinition{
	Name:                 "updatePassword",
	PrettyName:           "Update Password",
	Description:          "Updates the user's password in the database with a new hashed password",
	Category:             "User Management",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"user"},
	OutputContext:        []string{}, // or we may skip outputs if conditions imply it
	PossibleResultStates: []string{"success", "fail"},
	CustomConfigOptions: map[string]string{
		"minLength": "Minimum password length when setting a new password (default: 8)",
		"maxLength": "Maximum password length when setting a new password (default: 64)",
	},
	Run: RunUpdatePasswordNode,
}

func RunUpdatePasswordNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	if state.User == nil {
		return model.NewNodeResultWithError(errors.New("user must be loaded before updating password"))
	}

	if state.Context == nil {
		state.Context = map[string]string{}
	}

	// If the password is not set in the input, use the password from the context
	if password, submitted := input["password"]; submitted {
		state.Context["password"] = password
	}

	if state.Context["password"] == "" {
		return model.NewNodeResultWithPrompts(map[string]string{"password": "password"})
	}

	var customConfig map[string]string
	if node != nil {
		customConfig = node.CustomConfig
	}
	if err := lib.ValidateNewPassword(state.Context["password"], lib.PasswordLimitsForStore(customConfig)); err != nil {
		state.Error = lib.StringPtr(err.Error())
		return model.NewNodeResultWithPrompts(map[string]string{"password": "password"})
	}
	state.Error = nil

	hashed, err := lib.HashPassword(state.Context["password"])
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to hash password: %w", err))
	}

	// Create or update password attribute
	passwordAttrValue := model.PasswordAttributeValue{
		PasswordHash:         hashed,
		Locked:               false,
		FailedAttempts:       0,
		LastCorrectTimestamp: timePtr(time.Now()),
	}

	// Check if user already has a password attribute
	existingPasswordAttr, _, err := model.GetAttribute[model.PasswordAttributeValue](state.User, model.AttributeTypePassword)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to check existing password attribute: %w", err))
	}

	if existingPasswordAttr != nil {
		// Create new password attribute value
		newPasswordAttrValue := model.PasswordAttributeValue{
			PasswordHash:         hashed,
			Locked:               false,
			FailedAttempts:       0,
			LastCorrectTimestamp: timePtr(time.Now()),
		}

		// Find the attribute in UserAttributes and update it
		for i, attr := range state.User.UserAttributes {
			if attr.Type == model.AttributeTypePassword {
				state.User.UserAttributes[i].Value = newPasswordAttrValue
				break
			}
		}
	} else {
		// Create new password attribute
		passwordAttr := &model.UserAttribute{
			Type:  model.AttributeTypePassword,
			Value: passwordAttrValue,
		}
		state.User.AddAttribute(passwordAttr)
	}

	// Update the user in the database
	err = services.UserRepo.Update(context.Background(), state.User)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to update password: %w", err))
	}

	return model.NewNodeResultWithCondition("success")
}
