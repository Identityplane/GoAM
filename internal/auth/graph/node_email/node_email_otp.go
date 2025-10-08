package node_email

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	EMAIL_OTP_OPTION_MAX_ATTEMPTS = "max_attempts"
	EMAIL_OTP_OPTION_INIT_USER    = "init_user"
	EMAIL_OTP_OPTION_ASK_EMAIL    = "ask_email"

	MSG_INVALID_OTP = "Invalid OTP"
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
		EMAIL_OTP_OPTION_ASK_EMAIL:    "If true, the node will ask for an email before sending the OTP. Default false",
	},
	Run: RunEmailOTPNode,
}

func RunEmailOTPNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	askEmail := node.CustomConfig[EMAIL_OTP_OPTION_ASK_EMAIL] == "true"

	// If we ask for an email then ask for it if not in the context
	if askEmail {

		if input["email"] != "" {
			state.Context["email"] = input["email"]
		}

		if state.Context["email"] == "" {
			return model.NewNodeResultWithPrompts(map[string]string{"email": "email"})
		}
	}

	// Get the email address that we are using for this node
	email, user, error := getEmailAddress(state, services)
	if error != nil {
		return model.NewNodeResultWithError(error)
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

	// Input otp from the user
	otp := input["otp"]

	// OTP challenge stored on the server
	otpChallange := state.Context["email_otp"]

	// If we have no OTP challenge we generate a new one
	if otpChallange == "" {

		otpChallange := generateOTP()
		sendEmailOTP(email, otpChallange, user, services, mfa_max_attempts)
		state.Context["email_otp"] = otpChallange

		return otpPrompt(askEmail, email)
	} else if input["option"] == "resend" {

		sendEmailOTP(email, otpChallange, user, services, mfa_max_attempts)
		return otpPrompt(askEmail, email)
	} else if otp == "" {
		return otpPrompt(askEmail, email)
	}

	// If we have an opt we verify it
	isValid := (otp == state.Context["email_otp"])

	if !isValid {

		state.Context["error"] = MSG_INVALID_OTP
		err := increaseOTPFailedAttempts(email, user, services)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}

		// We ask again for the OTP but dont send a new email
		return otpPrompt(askEmail, email)
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
func sendEmailOTP(email string, otp string, user *model.User, services *model.Repositories, maxFailedAttempts int) error {

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

func otpPrompt(askEmail bool, email string) (*model.NodeResult, error) {
	if askEmail {
		return model.NewNodeResultWithPrompts(map[string]string{"otp": "number", "email": email})
	} else {
		return model.NewNodeResultWithPrompts(map[string]string{"otp": "number"})
	}
}
