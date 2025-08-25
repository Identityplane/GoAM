package node_passkeys

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

var PasskeyRegisterNode = &model.NodeDefinition{
	Name:            "registerPasskey",
	PrettyName:      "Register Passkey",
	Description:     "Registers a new passkey (WebAuthn credential) for the user to enable passwordless authentication. User must already be logged in",
	Category:        "Passkeys",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{"user"},
	PossiblePrompts: map[string]string{
		"passkeysFinishRegistrationJson": "json",
	},
	OutputContext:        []string{"passkeysFinishRegistrationJson", "passkeysSession", "passkeysOptions"},
	PossibleResultStates: []string{"success", "failure"},
	Run:                  RunPasskeyRegisterNode,
}

func RunPasskeyRegisterNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// User must be loaded before generating passkey options
	user, err := node_utils.LoadUserFromContext(state, services)
	if err != nil {
		return nil, fmt.Errorf("user must be loaded before registering a passkey: %w", err)
	}

	// We need the account name to generate the passkey options
	accountName := node_utils.GetAccountNameFromContext(state)

	// Check if input is present, if not generate options, if present process registration
	if _, ok := input["passkeysFinishRegistrationJson"]; !ok {

		// Generate options
		prompts, err := generatePasskeysOptions(state, node, accountName, user.ID)
		if err != nil {

			return model.NewNodeResultWithError(fmt.Errorf("failed to generate passkeys options: %w", err))
		}
		return model.NewNodeResultWithPrompts(prompts)

	} else {

		// Process registration
		result, err := ProcessPasskeyRegistration(state, node, input, services, accountName, user)
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to process passkey registration: %w", err))
		}
		return model.NewNodeResultWithCondition(result)
	}
}

func ProcessPasskeyRegistration(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories, accountName string, user *model.User) (string, error) {

	ctx := context.Background()

	// Unmarshal the WebAuthn session data from the context
	sessionJSON := state.Context["passkeysSession"]
	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Unmarshal the credential response from the client
	responseJSONStr := input["passkeysFinishRegistrationJson"]

	// Parse the credential response from the client
	parsedCredential, err := protocol.ParseCredentialCreationResponseBytes([]byte(responseJSONStr))
	if err != nil {
		return "failure", fmt.Errorf("failed to parse credential response: %w", err)
	}

	// Recreate the user object (must match the one used in BeginRegistration)
	userCredentials := &WebAuthnUserCredentials{
		Username: accountName,
		ID:       []byte(user.ID),
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

	credentialID := string(cred.ID)

	// Store new credential with user
	passkeyAttributeValue := model.PasskeyAttributeValue{
		RPID:               wconfig.RPID,
		DisplayName:        accountName,
		CredentialID:       credentialID,
		WebAuthnCredential: cred,
	}

	user.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypePasskey,
		Index: credentialID,
		Value: passkeyAttributeValue,
	})

	// Update user the user with the new credential
	services.UserRepo.CreateOrUpdate(ctx, user)

	// Log the successful registration
	log := logger.GetLogger()
	log.Info().Str("credential_id", string(cred.ID)).Msg("successfully registered credential")
	return "success", nil
}
