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

var EmailOTPNode = &model.NodeDefinition{
	Name:                 "emailOTP",
	PrettyName:           "Email OTP Verification",
	Description:          "Sends a one-time password via email and verifies the user's response for multi-factor authentication. If this email belongs to a user we increase the failed login attempts counter. If no user is found there is no limit to the number of attempts. In that case the node should be behind a captcha or similar.",
	Category:             "Multi-Factor Authentication",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{"email"},
	PossiblePrompts:      map[string]string{"otp": "number"},
	OutputContext:        []string{"emailOTP", "email_verified"},
	PossibleResultStates: []string{"success", "failure", "locked"},
	CustomConfigOptions: map[string]string{
		"smtp_server":       "The SMTP server address used to send OTP emails",
		"smtp_port":         "The port number for the SMTP server",
		"smtp_username":     "The username for authenticating with the SMTP server",
		"smtp_password":     "The password for authenticating with the SMTP server",
		"smtp_sender_email": "The email address that will appear as the sender of the OTP email",
	},
	Run: RunEmailOTPNode,
}

func RunEmailOTPNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load the user from the context, if we have a user we count the failed attempts
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// First we need to identiy the email address. For that we load the email from context
	// or with second priority from the user attributes
	email := state.Context["email"]
	if email == "" {

		if user != nil {

			emailValue, _, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
			if err != nil {
				return model.NewNodeResultWithError(errors.New("could not load email from user attributes"))
			}

			email = emailValue.Email
		}
	}

	// If we still don't have an email we fail
	if email == "" {
		return model.NewNodeResultWithError(errors.New("email must be provided before running this node"))
	}

	// Max attempts for the OTP
	mfa_max_attempts := 10
	if v, ok := node.CustomConfig["mfa_max_attempts"]; ok {
		mfa_max_attempts, _ = strconv.Atoi(v)
	}

	otp := input["otp"]
	if otp == "" {

		isLocked := false

		// If we have a have a user from the context and it has an email attibute we check if it is locked
		// if it is locked we dont send a otp and fail silently
		if user != nil {
			emailValue, _, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
			if err != nil {
				return model.NewNodeResultWithError(errors.New("could not load email from user attributes"))
			}

			if emailValue != nil && emailValue.OtpLocked {
				isLocked = true
			}
		}

		otp = generateOTP()

		if !isLocked {
			sendEmailOTP(email, otp, node, services)
		}

		state.Context["email_otp"] = otp
		return model.NewNodeResultWithPrompts(map[string]string{"otp": "number"})
	}

	// If we have an opt we verify it
	isValid := (otp == state.Context["email_otp"])

	if !isValid && user != nil {

		emailValue, attribute, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
		if err != nil {
			return model.NewNodeResultWithError(errors.New("could not load email from user attributes"))
		}

		// Create new email attribute value
		newEmailValue := model.EmailAttributeValue{
			Email:             emailValue.Email,
			Verified:          emailValue.Verified,
			VerifiedAt:        emailValue.VerifiedAt,
			OtpFailedAttempts: emailValue.OtpFailedAttempts + 1,
			OtpLocked:         emailValue.OtpFailedAttempts+1 >= mfa_max_attempts,
		}

		// Update the email attribute in the user's UserAttributes slice
		for i, attr := range user.UserAttributes {
			if attr.ID == attribute.ID {
				user.UserAttributes[i].Value = newEmailValue
				// Update the attribute reference to point to the updated one
				attribute = &user.UserAttributes[i]
				break
			}
		}

		// save the updated email attribute
		services.UserRepo.UpdateUserAttribute(context.Background(), attribute)
	}

	if isValid && user != nil {
		emailValue, attribute, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
		if err != nil {
			return model.NewNodeResultWithError(errors.New("could not load email from user attributes"))
		}

		// Create new email attribute value
		now := time.Now()
		newEmailValue := model.EmailAttributeValue{
			Email:             emailValue.Email,
			Verified:          true,
			VerifiedAt:        &now,
			OtpFailedAttempts: 0,
			OtpLocked:         false,
		}

		// Update the email attribute in the user's UserAttributes slice
		for i, attr := range user.UserAttributes {
			if attr.ID == attribute.ID {
				user.UserAttributes[i].Value = newEmailValue
				// Update the attribute reference to point to the updated one
				attribute = &user.UserAttributes[i]
				break
			}
		}

		// save the updated email attribute
		services.UserRepo.UpdateUserAttribute(context.Background(), attribute)
	}

	// if the otp is wrong we return the same otp prompt again but increase the mfa counter
	if !isValid {
		state.Context["error"] = "Invalid OTP"
		state.Context["email"] = email
		state.Context["email_verified"] = "true"
		return model.NewNodeResultWithPrompts(map[string]string{"otp": "number"})
	}

	return model.NewNodeResultWithCondition("success")
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
func sendEmailOTP(email string, otp string, node *model.GraphNode, services *model.Repositories) error {

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
