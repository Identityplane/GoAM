package graph

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/logger"
	"goiam/internal/model"
	"math/big"
	"strconv"
)

var EmailOTPNode = &NodeDefinition{
	Name:                 "emailOTP",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{"email"},
	PossiblePrompts:      map[string]string{"otp": "number"},
	OutputContext:        []string{"emailOTP"},
	PossibleResultStates: []string{"success", "failure", "locked"},
	Run:                  RunEmailOTPNode,
}

func RunEmailOTPNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	otp := input["otp"]
	email := state.Context["email"]

	mfa_max_attempts := 10
	if v, ok := node.CustomConfig["mfa_max_attempts"]; ok {
		mfa_max_attempts, _ = strconv.Atoi(v)
	}

	if email == "" {
		return model.NewNodeResultWithError(errors.New("email must be provided before running this node"))
	}

	// Load the user by email
	user, err := services.UserRepo.GetByEmail(context.Background(), email)
	if err != nil {
		return model.NewNodeResultWithError(errors.New("could not load user"))
	}

	if otp == "" {

		// if we cannot find the user we fail silently by returning the same otp prompt but we log the error
		if user != nil {

			if user.FailedLoginAttemptsMFA >= mfa_max_attempts {
				return model.NewNodeResultWithCondition("locked")
			}

			otp = generateOTP()
			sendEmailOTP(email, otp)
		}

		state.Context["email_otp"] = otp
		return model.NewNodeResultWithPrompts(map[string]string{"otp": "number"})
	}

	// If we have an opt we verify it
	if otp == state.Context["email_otp"] {

		user.FailedLoginAttemptsMFA = 0
		services.UserRepo.Update(context.Background(), user)
		state.User = user
		return model.NewNodeResultWithCondition("success")
	}

	// if the otp is wrong we return the same otp prompt again but increase the mfa counter
	user.FailedLoginAttemptsMFA++
	services.UserRepo.Update(context.Background(), user)
	state.Context["message"] = "Invalid OTP"
	return model.NewNodeResultWithPrompts(map[string]string{"otp": "number"})
}

func generateOTP() string {

	// Cryptographically securely generate a random 6 digit OTP

	otp, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return ""
	}

	// Convert the OTP to a string
	return fmt.Sprintf("%06d", otp)
}

func sendEmailOTP(email string, otp string) {

	// As a mock we just log the OTP for now
	logger.Info("OTP sent to %s: %s", email, otp)
}
