package node_totp

import (
	"context"
	"errors"
	"strconv"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/pquerna/otp/totp"
)

var TOTPVerifyNode = &model.NodeDefinition{
	Name:            "verifyTOTP",
	PrettyName:      "Verify TOTP",
	Description:     "Verify a TOTP code for 2FA",
	Category:        "MFA",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{"user"},
	PossiblePrompts: map[string]string{
		"totpVerification": "string",
	},
	OutputContext:        []string{""},
	PossibleResultStates: []string{model.ResultStateSuccess, model.ResultStateFail, model.ResultStateLocked, model.ResultStateNotFound},
	CustomConfigOptions: map[string]string{
		"max_failed_attempts": "Maximum number of failed attempts before locking the user (default: 10)",
	},
	Run: RunTOTPVerifyNode,
}

const (
	DEFAULT_MAX_FAILED_ATTEMPTS = 10
)

func RunTOTPVerifyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// This node needs a user in the context
	if state.User == nil {
		return nil, errors.New("user not found in context - register otp needs a user in the context")
	}

	// Get the max failed attempts
	maxFailedAttempts := DEFAULT_MAX_FAILED_ATTEMPTS
	if node.CustomConfig["max_failed_attempts"] != "" {
		var err error
		maxFailedAttempts, err = strconv.Atoi(node.CustomConfig["max_failed_attempts"])
		if err != nil {
			return nil, err
		}
	}

	// Validate the TOTP verification code
	totpValue, attribute, err := model.GetAttribute[model.TOTPAttributeValue](state.User, model.AttributeTypeTOTP)
	if err != nil {
		return nil, err
	}

	// if the user has no totp attirbute we return
	if totpValue == nil {
		return model.NewNodeResultWithCondition(model.ResultStateNotFound)
	}

	// If we have no input we ask for the verification code
	if input["totpVerification"] == "" {
		return model.NewNodeResultWithPrompts(map[string]string{
			"totpVerification": "string",
		})
	}

	// If the totp is locked we return a locked state
	if totpValue.Locked {
		return model.NewNodeResultWithCondition(model.ResultStateLocked)
	}

	// Validate the TOTP verification code
	valid := totp.Validate(input["totpVerification"], totpValue.SecretKey)
	if !valid {

		// Increment the failed attempts
		totpValue.FailedAttempts++

		if totpValue.FailedAttempts >= maxFailedAttempts {
			// Lock the TOTP
			totpValue.Locked = true
		}

		// Update the attribute
		attribute.Value = *totpValue

		// We need to immediately save the attribute
		err = services.UserRepo.UpdateUserAttribute(context.Background(), attribute)
		if err != nil {
			return nil, err
		}

		// Update the user's in-memory attributes to reflect the database changes
		for i := range state.User.UserAttributes {
			if state.User.UserAttributes[i].Type == model.AttributeTypeTOTP {
				state.User.UserAttributes[i].Value = *totpValue
				break
			}
		}

		errorMessage := "Invalid Code"
		state.Error = &errorMessage
		return model.NewNodeResultWithCondition(model.ResultStateFail)
	} else {
		// Reset the failed attempts
		totpValue.FailedAttempts = 0

		// Update the attribute
		attribute.Value = *totpValue

		// We need to save the attribute
		err = services.UserRepo.UpdateUserAttribute(context.Background(), attribute)
		if err != nil {
			return nil, err
		}

		// Update the user's in-memory attributes to reflect the database changes
		for i := range state.User.UserAttributes {
			if state.User.UserAttributes[i].Type == model.AttributeTypeTOTP {
				state.User.UserAttributes[i].Value = *totpValue
				break
			}
		}

		return model.NewNodeResultWithCondition(model.ResultStateSuccess)
	}
}
