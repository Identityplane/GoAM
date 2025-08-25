package node_passkeys

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

var PasskeysVerifyNode = &model.NodeDefinition{
	Name:            "verifyPasskey",
	PrettyName:      "Verify Passkey",
	Description:     "Verifies a passkey (WebAuthn credential) for passwordless authentication",
	Category:        "Passkeys",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{"username"},
	PossiblePrompts: map[string]string{
		"passkeysFinishLoginJson": "json",
		"passkeysLoginOptions":    "json",
	},
	OutputContext:        []string{"passkeysFinishLoginJson", "passkeysSession", "passkeysLoginOptions"},
	PossibleResultStates: []string{"success", "failure"},
	Run:                  RunPasskeyVerifyNode,
}

func RunPasskeyVerifyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// If we have a passkeysFinishLoginJson in the input we add it to the context
	passkeysFinishLoginJson, ok := input["passkeysFinishLoginJson"]
	if ok {
		state.Context["passkeysFinishLoginJson"] = passkeysFinishLoginJson
	}

	// If we have a passkeysFinishLoginJson in the context we process the passkey login
	if _, ok := state.Context["passkeysFinishLoginJson"]; ok {

		// Process assertion
		result, err := ProcessPasskeyLogin(state, node, input, services)
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to process passkey login: %w", err))
		}
		return model.NewNodeResultWithCondition(result)
	}

	// Otherwise we generate options and prompt for passkeysFinishLoginJson
	//prompts, err := GeneratePasskeysLoginOptions(state, node, services)
	// For passkey discovery we create a passkey challenge
	passkeysLoginOptions, err := GeneratePasskeysChallenge(state, node, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate passkey challenge: %w", err)
	}

	return model.NewNodeResultWithPrompts(map[string]string{"passkeysLoginOptions": passkeysLoginOptions})
}

func ProcessPasskeyLogin(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (string, error) {
	log := logger.GetLogger()

	// Load session from context
	sessionJSON := state.Context["passkeysSession"]
	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal login session: %w", err)
	}

	// Parse credential from the passkeysFinishLoginJson object
	responseJSONStr := state.Context["passkeysFinishLoginJson"]
	parsedCredential, err := protocol.ParseCredentialRequestResponseBytes([]byte(responseJSONStr))
	if err != nil {
		return "failure", fmt.Errorf("failed to parse credential assertion: %w", err)
	}

	credentialID := parsedCredential.ID
	log.Debug().Str("credential_id", credentialID).Msg("credential id")

	// Load user by the credential id
	user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypePasskey, credentialID)
	if err != nil {
		return "failure", fmt.Errorf("failed to load user by credential id: %w", err)
	}
	if user == nil {
		return "failure", fmt.Errorf("credential id %s not found", credentialID)
	}

	passkeys, attributes, err := model.GetAttributes[model.PasskeyAttributeValue](user, model.AttributeTypePasskey)
	if err != nil {
		return "failure", fmt.Errorf("failed to get passkeys: %w", err)
	}

	if len(passkeys) == 0 {
		return "failure", fmt.Errorf("user has no registered passkey")
	}

	// Loop through the passkeys array and find the one with the correct credential id
	var passkeyValue *model.PasskeyAttributeValue
	var passkeyAttribute *model.UserAttribute
	for i, passkey := range passkeys {
		if passkey.CredentialID == credentialID {
			passkeyValue = &passkey
			passkeyAttribute = attributes[i]
			break
		}
	}
	if passkeyValue == nil {
		return "failure", fmt.Errorf("passkey with credential id %s not found", credentialID)
	}

	userCredentials := &WebAuthnUserCredentials{
		Username:    passkeyValue.DisplayName,
		ID:          []byte(passkeyValue.CredentialID),
		Credentials: []webauthn.Credential{*passkeyValue.WebAuthnCredential},
	}

	// WebAuthn verify
	wconfig, err := getWebAuthnConfig(state, node)
	if err != nil {
		return "failure", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	webAuth, err := webauthn.New(wconfig)
	if err != nil {
		return "failure", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	// if the user id from the session is empty we overwrite it with the user id from the user object
	// This is needed for passkey discovery as we don't know the user id in that case during session creation
	if len(session.UserID) == 0 {
		session.UserID = []byte(passkeyValue.CredentialID)
	}

	// Validate the passkey login
	_, err = webAuth.ValidateLogin(userCredentials, session, parsedCredential)
	if err != nil {
		return "failure", fmt.Errorf("passkey login failed: %w", err)
	}

	// Update the last used at
	now := time.Now()
	passkeyValue.LastUsedAt = &now
	passkeyAttribute.Value = passkeyValue
	err = services.UserRepo.UpdateUserAttribute(context.Background(), passkeyAttribute)
	if err != nil {
		return "failure", fmt.Errorf("failed to update passkey attribute: %w", err)
	}

	// Set the user in the state
	state.User = user

	log.Debug().Str("user_id", user.ID).Msg("user successfully verified via passkey")
	return "success", nil
}
