package node_yubico

import (
	"context"
	"errors"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONFIG_SKIP_SAVE_USER      = "skipsaveuser"
	CONFIG_YUBICO_API_URL      = "yubikey-apiurl"
	CONFIG_YUBICO_CLIENT_ID    = "yubikey-clientid"
	CONFIG_YUBICO_API_KEY      = "yubikey-apikey"
	CONFIG_CHECK_YUBICO_UNIQUE = "yubikey-checkunique"
)

var YubicoCreateNode = &model.NodeDefinition{
	Name:            "createYubikeyOtp",
	PrettyName:      "Set Yubikey OTP",
	Description:     "Yubikey OTP. User must already be in the context",
	Category:        "MFA",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	PossiblePrompts: map[string]string{
		"yubikeyOtpVerification": "string",
	},
	OutputContext: []string{""},
	CustomConfigOptions: map[string]string{
		CONFIG_SKIP_SAVE_USER:      "If true, the user will not be saved to the database after the OTP is created and only the context will be updated",
		CONFIG_YUBICO_API_URL:      "The URL of the Yubikey API, defaults to https://api.yubico.com/wsapi/2.0/verify",
		CONFIG_YUBICO_CLIENT_ID:    "The Client ID of the Yubikey API",
		CONFIG_YUBICO_API_KEY:      "The API Key of the Yubikey API",
		CONFIG_CHECK_YUBICO_UNIQUE: "If true, the node will check if another user already has this yubikey registered. This is needed to sign-in with the yubikey as primary method.",
	},
	PossibleResultStates: []string{
		model.ResultStateSuccess,
		model.ResultStateFailure,
		model.ResultStateExisting,
	},
	Run: RunYubicoCreateNode,
}

func RunYubicoCreateNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load user from context
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If we have no user we return an error
	if user == nil {
		return model.NewNodeResultWithError(errors.New("user not found in context"))
	}

	// Validate the Yubikey OTP
	apiUrl := node.CustomConfig[CONFIG_YUBICO_API_URL]
	if apiUrl == "" {
		apiUrl = DEFAULT_YUBICO_API_URL
	}
	clientId := node.CustomConfig[CONFIG_YUBICO_CLIENT_ID]
	if clientId == "" {
		return nil, errors.New("yubikey client id is required")
	}
	apiKey := node.CustomConfig[CONFIG_YUBICO_API_KEY]
	if apiKey == "" {
		return nil, errors.New("yubikey api key is required")
	}

	// If we have no input we ask for the OTP
	if input["yubikeyOtpVerification"] == "" {
		return model.NewNodeResultWithPrompts(map[string]string{
			"yubikeyOtpVerification": "string",
		})
	}

	// Create the verifier instance
	verifier := getYubikeyVerifier(apiUrl, clientId, apiKey)

	// Verify the OTP
	publicId, valid, err := verifier.VerifyYubicoOtp(input["yubikeyOtpVerification"])
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If the OTP is not valid we return a failure
	if !valid {
		return model.NewNodeResultWithCondition(model.ResultStateFailure)
	}

	// We only create the index for this attribute if we ensured that it is unique
	var index *string

	// Check if any other user already has this yubikey
	if node.CustomConfig[CONFIG_CHECK_YUBICO_UNIQUE] == "true" {
		otherUser, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeYubico, publicId)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
		if otherUser != nil {
			return model.NewNodeResultWithCondition(model.ResultStateExisting)
		} else {
			index = &publicId
		}
	}

	// Create the yubikey attribute
	yubikeyAttribute := &model.UserAttribute{
		Type:  model.AttributeTypeYubico,
		Index: index,
		Value: model.YubicoAttributeValue{
			PublicID:       publicId,
			Locked:         false,
			FailedAttempts: 0,
		},
	}

	// Add the yubikey attribute to the user
	state.User.AddAttribute(yubikeyAttribute)

	// If we are saving the user we need to save it to the database
	if node.CustomConfig[CONFIG_SKIP_SAVE_USER] != "true" {
		err := services.UserRepo.CreateOrUpdate(context.Background(), state.User)
		if err != nil {
			return nil, err
		}
	}

	return model.NewNodeResultWithCondition(model.ResultStateSuccess)
}
