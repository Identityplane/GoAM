package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

var PasskeyRegisterNode = &NodeDefinition{
	Name:   "registerPasskey",
	Type:   NodeTypeQueryWithLogic,
	Inputs: []string{"username"},
	Prompts: map[string]string{
		"passkeysFinishRegistrationJson": "json",
	},
	Outputs:           []string{"passkeysFinishRegistrationJson", "passkeysSession", "passkeysOptions"},
	Conditions:        []string{"success", "failure"},
	GeneratePrompts:   GeneratePasskeysOptions,
	ProcessSubmission: ProcessPasskeyRegistration,
}

func GeneratePasskeysOptions(state *FlowState, node *GraphNode) (map[string]string, error) {
	//ctx := context.Background()
	username := state.Context["username"]

	// Init passkys
	wconfig := &webauthn.Config{
		RPDisplayName: "Go IAM",                          // Display Name for your site
		RPID:          "localhost",                       // Generally the FQDN for your site
		RPOrigins:     []string{"http://localhost:8080"}, // The origin URLs allowed for WebAuthn requests
	}

	webAuth, err := webauthn.New(wconfig)

	if err != nil {
		panic(err)
	}

	user := &WebAuthnUserCredentials{
		Username: username,
		ID:       []byte(username),
	}
	options, session, err := webAuth.BeginRegistration(user)

	if err != nil {
		return nil, fmt.Errorf("failed to start passkeys registartion %w", err)
	}

	// Marshal session and options into JSON strings
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	state.Context["passkeysSession"] = string(sessionJSON)
	state.Context["passkeysOptions"] = string(optionsJSON)

	prompts := &map[string]string{"passkeysOptions": string(optionsJSON)}

	return *prompts, nil
}

func ProcessPasskeyRegistration(state *FlowState, node *GraphNode, input map[string]string) (string, error) {

	ctx := context.Background()

	// Unmarshal the WebAuthn session data from the context
	sessionJSON := state.Context["passkeysSession"]
	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Unmarshal the credential response from the client
	responseJSONStr := input["passkeysFinishRegistrationJson"]

	parsedCredential, err := protocol.ParseCredentialCreationResponseBytes([]byte(responseJSONStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse credential response: %w", err)
	}

	// Recreate the user object (must match the one used in BeginRegistration)
	username := state.Context["username"]
	user := &WebAuthnUserCredentials{
		Username: username,
		ID:       []byte(username),
	}

	// Re-initialize the WebAuthn config
	// TODO abstract this part
	wconfig := &webauthn.Config{
		RPDisplayName: "Go IAM",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}

	webAuth, err := webauthn.New(wconfig)
	if err != nil {
		return "failure", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	// Finish registration and store the credential if valid
	cred, err := webAuth.CreateCredential(user, session, parsedCredential)
	if err != nil {
		return "failure", fmt.Errorf("failed to finish registration: %w", err)
	}

	// Store new credential with user
	// First load
	userRepo := Services.UserRepo
	if userRepo == nil {
		return "fail", errors.New("UserRepo not initialized")
	}

	userModel, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return "fail", errors.New("Could not load user")
	}

	credBytes, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}
	userModel.Attributes["webauthn_credential"] = string(credBytes)

	// Then store again
	userRepo.Update(ctx, userModel)

	// Also store user into flow context
	userBytes, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}
	state.Context["user"] = string(userBytes)

	log.Printf("Successfully registered credential ID: %s", cred.ID)

	return "success", nil
}

// PasskeysBeginLoginNode

// PasskeysFinishLoginNode

// WebAuthnUserCredentials is a simple struct that implements the webauthn.User interface
type WebAuthnUserCredentials struct {
	ID          []byte
	Username    string
	DisplayName string
	Credentials []webauthn.Credential
}

// WebAuthnID returns the user's unique WebAuthn ID (opaque byte slice)
func (u *WebAuthnUserCredentials) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName returns the human-readable username
func (u *WebAuthnUserCredentials) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName returns the display name (for UI)
func (u *WebAuthnUserCredentials) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials returns the list of credentials registered with this user
func (u *WebAuthnUserCredentials) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
