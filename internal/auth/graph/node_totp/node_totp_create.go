package node_totp

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image/png"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var TOTPCreateNode = &model.NodeDefinition{
	Name:            "createTOTP",
	PrettyName:      "Set TOTP",
	Description:     "RFC 6238 compliant TOTP for 2FA. User must already be in the context",
	Category:        "MFA",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	PossiblePrompts: map[string]string{
		"totpVerification": "string",
	},
	OutputContext: []string{""},
	CustomConfigOptions: map[string]string{
		"totpIssuer":   "The name of the issuer displayed in the TOTP QR code",
		"skipSaveUser": "If true, the user will not be saved to the database after the TOTP is created and only the context will be updated",
	},
	PossibleResultStates: []string{"success"},
	Run:                  RunTOTPCreateNode,
}

func RunTOTPCreateNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// This node needs a user in the context
	if state.User == nil {
		return nil, errors.New("user not found in context - register otp needs a user in the context")
	}

	if input["totpVerification"] == "" {

		issuer := node.CustomConfig["totpIssuer"]
		if issuer == "" {
			issuer = "GoAM"
		}

		// Get the display name of the user
		accountName := node_utils.GetAccountNameFromContext(state)

		// If the account name is empty we set it to the issuer as a fallback because the totp library requires a non empty account name
		if accountName == "" {
			accountName = issuer
		}

		// Create a new TOTP secret
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      issuer,
			AccountName: accountName,
		})

		// If we cannot generate a TOTP secret we return an error
		if err != nil {
			return nil, err
		}

		// If we cannot generate a TOTP image we return an error
		imageUrl, err := getTotpImageUrl(*key)
		if err != nil {
			return nil, err
		}

		// Get the secret from the key
		secret := key.Secret()

		// Store the TOTP secret in the context
		state.Context["totpSecret"] = secret
		state.Context["totpImageUrl"] = imageUrl
		state.Context["totpIssuer"] = issuer

		// Return a prompt with the information for the user to scan the QR code or client to display the secret
		return model.NewNodeResultWithPrompts(map[string]string{
			"totpSecret":       secret,
			"totpVerification": "string",
			"totpIssuer":       issuer,
			"totpImageUrl":     imageUrl,
		})
	}

	// Validate the TOTP verification code
	verificationCode := input["totpVerification"]
	totpSecret := state.Context["totpSecret"]

	valid := totp.Validate(verificationCode, totpSecret)
	if !valid {

		errMsg := "Invalid Code"
		state.Error = &errMsg

		secret := state.Context["totpSecret"]
		issuer := state.Context["totpIssuer"]
		imageUrl := state.Context["totpImageUrl"]

		// During registration we dont need to increment the failed attempts
		// as we are not storing the totp secret in the database
		// so we can just prompt again and the user can try again
		return model.NewNodeResultWithPrompts(map[string]string{
			"totpSecret":       secret,
			"totpVerification": "string",
			"totpIssuer":       issuer,
			"totpImageUrl":     imageUrl,
		})
	}

	// Create the totp attribute
	totpAttribute := &model.UserAttribute{
		Type: model.AttributeTypeTOTP,
		Value: model.TOTPAttributeValue{
			SecretKey:      totpSecret,
			Locked:         false,
			FailedAttempts: 0,
		},
	}

	// Add the totp attribute to the user
	state.User.AddAttribute(totpAttribute)

	// If we are saving the user we need to save it to the database
	if node.CustomConfig["skipSaveUser"] != "true" {
		err := services.UserRepo.CreateOrUpdate(context.Background(), state.User)
		if err != nil {
			return nil, err
		}
	}

	return &model.NodeResult{
		Prompts:   nil,
		Condition: "success",
	}, nil
}

func getTotpImageUrl(key otp.Key) (string, error) {

	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}
	png.Encode(&buf, img)

	// Convert the buffer to a base64 url
	base64Url := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + base64Url, nil
}
