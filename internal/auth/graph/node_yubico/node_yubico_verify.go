package node_yubico

import (
	"context"
	"errors"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONFIG_YUBICO_CREATE_USER = "yubikey-createuserifnotfound"
)

var YubicoVerifyNode = &model.NodeDefinition{
	Name:            "verifyYubikeyOtp",
	PrettyName:      "Verify Yubikey OTP",
	Description:     "Verify Yubikey OTP. If indexes are used user can be loaded via the otp verification. As yubikeys are considered a not brute foracble method invalid attempts are not storer and the user not locked.",
	Category:        "MFA",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	PossiblePrompts: map[string]string{
		"yubikeyOtpVerification": "string",
	},
	OutputContext: []string{""},
	CustomConfigOptions: map[string]string{
		CONFIG_YUBICO_API_URL:      "The URL of the Yubikey API, defaults to https://api.yubico.com/wsapi/2.0/verify",
		CONFIG_YUBICO_CLIENT_ID:    "The Client ID of the Yubikey API",
		CONFIG_YUBICO_API_KEY:      "The API Key of the Yubikey API",
		CONFIG_CHECK_YUBICO_UNIQUE: "If true, the node will check if another user already has this yubikey registered. This is needed to sign-in with the yubikey as primary method.",
		CONFIG_YUBICO_CREATE_USER:  "If true, the node will create a user if the user is not found in the context. This is only allowed if yubikey-checkunique is true.",
	},
	PossibleResultStates: []string{
		model.ResultStateSuccess,
		model.ResultStateFailure,
		model.ResultStateLocked,
		model.ResultStateNotFound,
		model.ResultStateNewUser,
	},
	Run: RunYubicoVerifyNode,
}

var log = logger.GetGoamLogger()

func RunYubicoVerifyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Try to load user from context
	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	log.Debug().Bool(CONFIG_CHECK_YUBICO_UNIQUE, node.CustomConfig[CONFIG_CHECK_YUBICO_UNIQUE] == "true").Msg("Running Yubico Verify Node")

	// This node needs a user in the context
	if user == nil && node.CustomConfig[CONFIG_CHECK_YUBICO_UNIQUE] != "true" {
		return nil, errors.New("user not found in context - register otp needs a user in the context")
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

		state.Error = &[]string{"Invalid Yubikey OTP"}[0]
		return model.NewNodeResultWithCondition(model.ResultStateFailure)
	}

	// If we dont have a user yet in the context and we use unique yubikeys we load the user from the database
	if state.User == nil && node.CustomConfig[CONFIG_CHECK_YUBICO_UNIQUE] == "true" {
		user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeYubico, publicId)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
		if user == nil {

			if node.CustomConfig[CONFIG_YUBICO_CREATE_USER] == "true" {

				// Create a new user
				err = CreateUserWithYubico(services, state, node, publicId)
				if err != nil {
					return model.NewNodeResultWithError(err)
				}
				return model.NewNodeResultWithCondition(model.ResultStateNewUser)

			} else {
				return model.NewNodeResultWithCondition(model.ResultStateNotFound)
			}
		}
		state.User = user
	}

	// Check that the user has a yubikey attribute with the name publicId
	yubikeyAttributes, _, err := model.GetAttributes[model.YubicoAttributeValue](state.User, model.AttributeTypeYubico)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	for _, attr := range yubikeyAttributes {

		if attr.PublicID == publicId {

			if attr.Locked {
				return model.NewNodeResultWithCondition(model.ResultStateLocked)
			}

			return model.NewNodeResultWithCondition(model.ResultStateSuccess)
		}
	}
	return model.NewNodeResultWithCondition(model.ResultStateFailure)
}

func CreateUserWithYubico(services *model.Repositories, state *model.AuthenticationSession, node *model.GraphNode, publicId string) error {

	var index *string
	var err error

	if node.CustomConfig[CONFIG_CHECK_YUBICO_UNIQUE] == "true" {
		index = &publicId
	}

	// Create the yubikey attribute
	yubikeyAttribute := &model.UserAttribute{
		Type:  model.AttributeTypeYubico,
		Index: index,
		Value: model.YubicoAttributeValue{
			PublicID: publicId,
			CredentialAttributeValue: model.CredentialAttributeValue{
				Locked:         false,
				FailedAttempts: 0,
			},
		},
	}

	if state.User == nil {
		state.User, err = services.UserRepo.NewUserModel(state)
		if err != nil {
			return err
		}
	}

	// Add the yubikey attribute to the user
	state.User.AddAttribute(yubikeyAttribute)

	// If we are saving the user we need to save it to the database
	if node.CustomConfig[CONFIG_SKIP_SAVE_USER] != "true" {
		err := services.UserRepo.Create(context.Background(), state.User)
		if err != nil {
			return err
		}
	}

	return nil
}
