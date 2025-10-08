package node_email

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	EMAIL_OTP_OPTION_MAX_ATTEMPTS = "max_attempts"
	EMAIL_OTP_OPTION_INIT_USER    = "init_user"
	EMAIL_RESEND_IN_SECONDS       = "resend_in_seconds"

	MSG_INVALID_OTP     = "Invalid OTP"
	MSG_RESEND_TOO_SOON = "You cannot resent the OTP yet"
)

var EmailOTPNode = &model.NodeDefinition{
	Name:                 "emailOTP",
	PrettyName:           "Email OTP Verification",
	Description:          "Sends a one-time password via email and verifies the user's response for multi-factor authentication. If this email belongs to a user we increase the failed login attempts counter. If no user is found there is no limit to the number of attempts. In that case the node should be behind a captcha or similar.",
	Category:             "Multi-Factor Authentication",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{"email"},
	PossiblePrompts:      map[string]string{"otp": "number", "option": "resend", "email": "email"},
	OutputContext:        []string{"emailOTP", "email_verified"},
	PossibleResultStates: []string{"success-registered-email", "success-unkown-email"},
	CustomConfigOptions: map[string]string{
		EMAIL_OTP_OPTION_MAX_ATTEMPTS: "Maximum number of failed attempts before locking the user (default: 10)",
		EMAIL_OTP_OPTION_INIT_USER:    "If true, the node will initialize the user if no user is found in the context. Default false",
		EMAIL_RESEND_IN_SECONDS:       "The number of seconds to wait before resending the OTP. Default 30",
	},
	Run: RunEmailOTPNode,
}

func RunEmailOTPNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Get the email address that we are using for this node
	email, user, err := getEmailAddress(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we don't have an email we fail
	if email == "" {
		return model.NewNodeResultWithError(errors.New("email must be provided before running this node"))
	}

	// Max attempts for the OTP
	mfa_max_attempts := 10
	if v, ok := node.CustomConfig[EMAIL_OTP_OPTION_MAX_ATTEMPTS]; ok {
		mfa_max_attempts, _ = strconv.Atoi(v)
	}
	resendInSeconds := 30
	if node.CustomConfig[EMAIL_RESEND_IN_SECONDS] != "" {

		var err error
		resendInSeconds, err = strconv.Atoi(node.CustomConfig[EMAIL_RESEND_IN_SECONDS])
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
	}

	// Input otp from the user
	otp := input["otp"]

	// OTP challenge stored on the server
	otpChallange := state.Context["email_otp"]

	// If we have no OTP challenge we generate a new one
	if otpChallange == "" {

		otpChallange := generateOTP()
		sendEmailOTP(email, otpChallange, user, services, mfa_max_attempts, state, resendInSeconds)
		state.Context["email_otp"] = otpChallange

		return otpPrompt(email, state)
	} else if input["option"] == "resend" {

		resendAt, err := time.Parse(time.RFC3339, state.Context["resend_at"])

		if err != nil || time.Now().After(resendAt) {

			sendEmailOTP(email, otpChallange, user, services, mfa_max_attempts, state, resendInSeconds)
			state.Context["message"] = ""
		} else {
			state.Context["message"] = MSG_RESEND_TOO_SOON
		}

		return otpPrompt(email, state)
	} else if otp == "" {
		return otpPrompt(email, state)
	}

	// If we have an opt we verify it
	isValid := (otp == state.Context["email_otp"])

	if !isValid {

		state.Context["message"] = MSG_INVALID_OTP
		err := increaseOTPFailedAttempts(email, user, services)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}

		// We ask again for the OTP but dont send a new email
		return otpPrompt(email, state)
	} else {

		err := registerSucessfullVerification(email, user, services)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}

		if user != nil {
			return model.NewNodeResultWithCondition("success-registered-email")
		} else {

			if node.CustomConfig[EMAIL_OTP_OPTION_INIT_USER] == "true" {
				user, err := services.UserRepo.NewUserModel(state)
				if err != nil {
					return model.NewNodeResultWithError(err)
				}

				// add a new email attribute with the verified email
				now := time.Now()
				emailAttribute := &model.UserAttribute{
					Type:  model.AttributeTypeEmail,
					Index: lib.StringPtr(email),
					Value: model.EmailAttributeValue{
						Email:      email,
						Verified:   true,
						VerifiedAt: &now,
					},
				}
				user.AddAttribute(emailAttribute)

				state.User = user
			}

			return model.NewNodeResultWithCondition("success-unkown-email")
		}

	}
}

// generateOTP generates a random 6 digit OTP
func generateOTP() string {

	// Cryptographically securely generate a random 6 digit OTP
	otp, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return ""
	}

	// Convert the OTP to a string
	return fmt.Sprintf("%06d", otp)
}

// sendEmailOTP sends an email with the OTP to the email address
func sendEmailOTP(email string, otp string, user *model.User, services *model.Repositories, maxFailedAttempts int, state *model.AuthenticationSession, resendInSeconds int) error {

	if user != nil {
		// Check if the email attribute is locked or the maximum number of failed attempts is reached
		emailValue, _, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
		if err != nil {
			return err
		}

		if emailValue.OtpLocked || emailValue.OtpFailedAttempts >= maxFailedAttempts {
			// Silently return but log the attempt
			return nil
		}
	}

	emailParams := &model.SendEmailParams{
		Template: "email-otp",
		To: []model.EmailAddress{
			{Email: email},
		},
		Params: map[string]any{
			"otp": otp,
		},
	}

	resendAt := time.Now().Add(time.Duration(resendInSeconds) * time.Second)
	state.Context["resend_at"] = resendAt.Format(time.RFC3339)

	services.EmailSender.SendEmail(emailParams)

	return nil
}

func getEmailAddress(state *model.AuthenticationSession, services *model.Repositories) (string, *model.User, error) {
	// First we need to identiy the email address. For that we load the email from context
	// or with second priority from the user attributes

	// Try to load the user from the context, if we have a user we count the failed attempts
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return "", nil, err
	}

	email := state.Context["email"]
	if email == "" {

		if user != nil {

			emailValue, _, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
			if err != nil {
				return "", nil, errors.New("could not load email from user attributes")
			}

			email = emailValue.Email
		}
	}

	return email, user, nil
}

func increaseOTPFailedAttempts(email string, user *model.User, services *model.Repositories) error {

	if user != nil {
		// If this email is registered to a user as primary email we increase the failed attempts for the attribute

		emailValue, attribute, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
		if err != nil {
			return err
		}

		emailValue.OtpFailedAttempts++
		attribute.Value = emailValue

		err = services.UserRepo.UpdateUserAttribute(context.Background(), attribute)
		if err != nil {
			return err
		}

		return nil
	} else {

		// Otherwise we increase the failed attempts for this email
		// TODO implement
		return nil
	}
}

func registerSucessfullVerification(email string, user *model.User, services *model.Repositories) error {

	now := time.Now()

	if user != nil {
		// If this email is registered to a user we set the verified flag to true and reset the number of failed attempts

		emailValue, attribute, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
		if err != nil {
			return err
		}

		emailValue.OtpFailedAttempts = 0
		emailValue.OtpLocked = false

		if !emailValue.Verified {
			emailValue.Verified = true
			emailValue.VerifiedAt = &now
		}

		attribute.Value = emailValue

		err = services.UserRepo.UpdateUserAttribute(context.Background(), attribute)
		if err != nil {
			return err
		}

		return nil
	} else {

		// If this does not belong to a user we reset the failed attempts for this email
		// TODO implement
		return nil
	}
}

func otpPrompt(email string, state *model.AuthenticationSession) (*model.NodeResult, error) {

	// Check how long the user needs to wait until they can request to resent the otp
	resendAt, err := time.Parse(time.RFC3339, state.Context["resend_at"])
	if err != nil {
		resendAt = time.Now()
	}

	secondToResend := int(math.Max(0, time.Until(resendAt).Seconds()))

	return model.NewNodeResultWithPrompts(map[string]string{"otp": "number", "email": email, "resend_in_seconds": strconv.Itoa(secondToResend)})
}
