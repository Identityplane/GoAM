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
	CustomConfigOptions:  []string{"smtp_server", "smtp_port", "smtp_username", "smtp_password", "smtp_sender_email"},
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
			sendEmailOTP(email, otp, node, services)
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

func sendEmailOTP(email string, otp string, node *model.GraphNode, services *repository.Repositories) error {

	// As a mock we just log the OTP for now
	logger.Info("OTP sent to %s: %s", email, otp)

	smtpServer := node.CustomConfig["smtp_server"]
	smtpPort := node.CustomConfig["smtp_port"]
	smtpUsername := node.CustomConfig["smtp_username"]
	smtpPassword := node.CustomConfig["smtp_password"]
	smtpSenderEmail := node.CustomConfig["smtp_sender_email"]

	if smtpServer == "" || smtpPort == "" || smtpUsername == "" || smtpPassword == "" || smtpSenderEmail == "" {
		msg := "smtp server, port, username, password, and sender email must be provided in the custom config. Otherwise the email will not be sent and fail silently."
		logger.ErrorNoContext(msg)
		return nil
	}

	body, subject, err := generateEmailBody(otp)
	if err != nil {
		msg := "error generating email body: %s"
		logger.ErrorNoContext(msg, err)
		return errors.New(msg)
	}

	err = services.EmailSender.SendEmail(subject, body, email, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail)
	if err != nil {
		msg := "error sending email: %s"
		logger.ErrorNoContext(msg, err)
		return errors.New(msg)
	}

	return nil
}

func generateEmailBody(otp string) (string, string, error) {

	body := fmt.Sprintf(defaultEmailBodyTemplate, otp)
	subject := defaultEmailSubject

	return body, subject, nil
}

const defaultEmailBodyTemplate = `Subject: Verify your identity with OTP
Please use the verification code below to confirm your identity.


Verification code:

%s`

const defaultEmailSubject = "Verify your identity"
