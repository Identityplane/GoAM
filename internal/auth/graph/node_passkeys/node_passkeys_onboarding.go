package node_passkeys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

var PasskeyOnboardingNode = &model.NodeDefinition{
	Name:            "onboardingWithPasskey",
	PrettyName:      "Passkey Onboarding",
	Description:     "Onboards a new user by asking for their email and then registering a passkey. No user will be created if the user asks for a password instead.",
	Category:        "Passkeys",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{"username"},
	PossiblePrompts: map[string]string{
		"passkeysFinishRegistrationJson": "json",
		"email":                          "email",
		"option":                         "text",
	},
	CustomConfigOptions: map[string]string{
		"createUser":         "If set to 'true' the user will be created in the database, otherwise the user will only be store in the context. Enable this option if you want to store the user immediately in the database. If you want to do additional processsing before such as validating the email you should not enable this option and manually create the user in a following node.",
		"showChoosePassword": "If set to 'true' the user has the option to choose a password instead of a passkey.",
	},
	OutputContext:        []string{"passkeysFinishRegistrationJson", "passkeysSession", "passkeysOptions"},
	PossibleResultStates: []string{"success", "failure", "choosesPassword", "existing"},
	Run:                  RunPasskeyOnboardingNode,
}

func RunPasskeyOnboardingNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Check if input is present, if not generate options, if present process registration
	if _, ok := input["option"]; !ok {

		// Generate options
		prompts, err := generateAnonymousPasskeysOptions(state, node)
		if err != nil {

			return model.NewNodeResultWithError(fmt.Errorf("failed to generate passkeys options: %w", err))
		}
		return model.NewNodeResultWithPrompts(prompts)

	} else {

		// check which option the user chose with the action parameter
		action := input["option"]
		switch action {
		case "passkey":

			// Process registration
			result, err := processPasskeyOnboarding(state, node, input, services)
			if err != nil {
				return model.NewNodeResultWithError(fmt.Errorf("failed to process passkey onboarding: %w", err))
			}

			return model.NewNodeResultWithCondition(result)
		case "password":

			if node.CustomConfig["showChoosePassword"] == "true" {
				return model.NewNodeResultWithCondition("choosesPassword")
			} else {
				return model.NewNodeResultWithError(fmt.Errorf("passsword is not available"))
			}
		default:
			return model.NewNodeResultWithError(fmt.Errorf("invalid action: %s", action))
		}
	}
}

func generateAnonymousPasskeysOptions(state *model.AuthenticationSession, node *model.GraphNode) (map[string]string, error) {

	// generate a random user id
	newUserId := uuid.New().String()

	// Store the new user id in the context
	state.Context["newUserId"] = newUserId

	optionsJSON, err := GeneratePasskeysChallenge(state, node, "", newUserId)

	if err != nil {
		return nil, fmt.Errorf("failed to generate passkeys challenge: %w", err)
	}

	prompts := &map[string]string{"passkeysOptions": string(optionsJSON)}

	return *prompts, nil
}

func processPasskeyOnboarding(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (string, error) {

	ctx := context.Background()

	// get the email from the input
	email := input["email"]
	if email == "" {
		return "failure", fmt.Errorf("email is required")
	}

	// Check if userRepo is initialized
	userRepo := services.UserRepo
	if userRepo == nil {
		return "failure", errors.New("userRepo not initialized")
	}

	// Check for existing user
	existing, _ := userRepo.GetByAttributeIndex(ctx, model.AttributeTypeEmail, email)
	if existing != nil {
		return "existing", nil
	}

	// Unmarshal the WebAuthn session data from the context
	sessionJSON := state.Context["passkeysSession"]
	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Unmarshal the credential response from the client
	responseJSONStr := input["passkeysFinishRegistrationJson"]

	// Parse the credential response
	parsedCredential, err := protocol.ParseCredentialCreationResponseBytes([]byte(responseJSONStr))
	if err != nil {
		return "failure", fmt.Errorf("failed to parse credential response: %w", err)
	}

	// Recreate the user object based on the email that the user chooses
	newUserId := state.Context["newUserId"]

	// Create the user credentials
	userCredentials := &WebAuthnUserCredentials{
		Username: email,
		ID:       []byte(newUserId),
	}

	// Re-initialize the WebAuthn config
	wconfig, err := getWebAuthnConfig(state, node)
	if err != nil {
		return "failure", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}
	webAuth, err := webauthn.New(wconfig)
	if err != nil {
		return "failure", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	// Finish registration and store the credential if valid
	cred, err := webAuth.CreateCredential(userCredentials, session, parsedCredential)
	if err != nil {
		return "failure", fmt.Errorf("failed to finish registration: %w", err)
	}

	// Create the passkey attribute
	passkeyAttribute := &model.UserAttribute{
		Type:  model.AttributeTypePasskey,
		Index: lib.StringPtr(newUserId),
		Value: model.PasskeyAttributeValue{
			CredentialID:       credIdToString(cred.ID),
			DisplayName:        email,
			RPID:               wconfig.RPID,
			WebAuthnCredential: cred,
		},
	}

	// Create the email attribute
	emailAttribute := &model.UserAttribute{
		Type:  model.AttributeTypeEmail,
		Index: lib.StringPtr(newUserId),
		Value: model.EmailAttributeValue{
			Email:    email,
			Verified: false,
		},
	}

	// Create a new user object with the provided email and previously generated user id
	user := &model.User{
		ID:        newUserId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add the attributes to the user
	user.AddAttribute(passkeyAttribute)
	user.AddAttribute(emailAttribute)

	// Set the user to the context
	state.User = user

	// If create user is set in the custom config then store the user with updated credential
	if node.CustomConfig["createUser"] == "true" {

		// Store the user in the database
		err = services.UserRepo.Create(ctx, user)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
	}

	log := logger.GetGoamLogger()
	log.Info().Str("credential_id", string(cred.ID)).Msg("successfully registered credential")
	return "success", nil
}
