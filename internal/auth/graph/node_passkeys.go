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
	Name:            "registerPasskey",
	Type:            NodeTypeQueryWithLogic,
	RequiredContext: []string{"username"},
	Prompts: map[string]string{
		"passkeysFinishRegistrationJson": "json",
	},
	OutputContext: []string{"passkeysFinishRegistrationJson", "passkeysSession", "passkeysOptions"},
	Conditions:    []string{"success", "failure"},
	Run:           RunPasskeyRegisterNode,
}

var PasskeysVerifyNode = &NodeDefinition{
	Name:            "verifyPasskey",
	Type:            NodeTypeQueryWithLogic,
	RequiredContext: []string{"username"},
	Prompts: map[string]string{
		"passkeysFinishLoginJson": "json",
	},
	OutputContext: []string{"passkeysFinishLoginJson", "passkeysLoginSession", "passkeysLoginOptions"},
	Conditions:    []string{"success", "failure"},
	Run:           RunPasskeyVerifyNode,
}

var PasskeysCheckUserRegistered = &NodeDefinition{
	Name:            "checkPasskeyRegistered",
	Type:            NodeTypeLogic,
	RequiredContext: []string{"username"},
	Prompts:         nil,
	OutputContext:   []string{},
	Conditions:      []string{"registered", "not_registered", "user_not_found"},
	Run:             RunCheckUserHasPasskeyNode,
}

func RunCheckUserHasPasskeyNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {
	ctx := context.Background()
	username := state.Context["username"]

	// Load user from DB
	userModel, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil || userModel == nil {
		return NewNodeResultWithCondition("user_not_found")
	}

	// Check if the user has a passkey registered
	_, ok := userModel.Attributes["webauthn_credential"]
	if !ok {
		state.Context["hasPasskeyRegistered"] = "not_registered"
		return NewNodeResultWithCondition("not_registered")
	} else {
		state.Context["hasPasskeyRegistered"] = "registered"
		return NewNodeResultWithCondition("registered")
	}
}

func RunPasskeyRegisterNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {

	// Check if input is present, if not generate options, if present process registration
	if _, ok := input["passkeysFinishRegistrationJson"]; !ok {

		// Generate options
		prompts, err := GeneratePasskeysOptions(state, node)
		if err != nil {

			return NewNodeResultWithError(fmt.Errorf("failed to generate passkeys options: %w", err))
		}
		return NewNodeResultWithPrompts(prompts)

	} else {
		// Process registration
		result, err := ProcessPasskeyRegistration(state, node, input)
		if err != nil {
			return NewNodeResultWithError(fmt.Errorf("failed to process passkey registration: %w", err))
		}
		return NewNodeResultWithCondition(result)
	}
}

func RunPasskeyVerifyNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {

	// Check if input is present, if not generate options, if present process assertion
	if _, ok := input["passkeysFinishLoginJson"]; !ok {

		// Generate options
		prompts, err := GeneratePasskeysLoginOptions(state, node)
		if err != nil {
			return NewNodeResultWithError(fmt.Errorf("failed to generate passkeys options: %w", err))
		}
		return NewNodeResultWithPrompts(prompts)

	} else {
		// Process assertion
		result, err := ProcessPasskeyLogin(state, node, input)
		if err != nil {
			return NewNodeResultWithError(fmt.Errorf("failed to process passkey login: %w", err))
		}
		return NewNodeResultWithCondition(result)
	}
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
		return "failure", fmt.Errorf("failed to parse credential response: %w", err)
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
		return "fail", errors.New("userRepo not initialized")
	}

	userModel, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return "fail", errors.New("could not load user")
	}

	credBytes, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}
	userModel.Attributes["webauthn_credential"] = string(credBytes)

	// Then store again
	err = userRepo.Update(ctx, userModel)
	if err != nil {
		return "", fmt.Errorf("failed to update user: %w", err)
	}

	// Also store user into flow context
	userBytes, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}
	state.Context["user"] = string(userBytes)

	log.Printf("Successfully registered credential ID: %s", cred.ID)

	return "success", nil
}

func GeneratePasskeysLoginOptions(state *FlowState, node *GraphNode) (map[string]string, error) {
	username := state.Context["username"]

	// Setup config
	wconfig := &webauthn.Config{
		RPDisplayName: "Go IAM",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}

	webAuth, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webauthn config: %w", err)
	}

	ctx := context.Background()
	userModel, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil || userModel == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	credJSON, ok := userModel.Attributes["webauthn_credential"]
	if !ok || len(credJSON) == 0 {
		return nil, fmt.Errorf("user has no registered passkey")
	}

	var cred webauthn.Credential
	if err := json.Unmarshal([]byte(credJSON), &cred); err != nil {
		return nil, fmt.Errorf("failed to parse stored credential: %w", err)
	}

	user := &WebAuthnUserCredentials{
		Username:    userModel.Username,
		ID:          []byte(userModel.Username),
		Credentials: []webauthn.Credential{cred},
	}

	options, session, err := webAuth.BeginLogin(user)
	if err != nil {
		return nil, fmt.Errorf("failed to start passkey login: %w", err)
	}

	// Store in context
	sessionJSON, _ := json.Marshal(session)
	optionsJSON, _ := json.Marshal(options)
	state.Context["passkeysLoginSession"] = string(sessionJSON)
	state.Context["passkeysLoginOptions"] = string(optionsJSON)

	return map[string]string{
		"passkeysLoginOptions": string(optionsJSON),
	}, nil
}

func ProcessPasskeyLogin(state *FlowState, node *GraphNode, input map[string]string) (string, error) {
	ctx := context.Background()
	username := state.Context["username"]

	// Load session from context
	sessionJSON := state.Context["passkeysLoginSession"]
	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal login session: %w", err)
	}

	// Parse credential from user input
	responseJSONStr := input["passkeysFinishLoginJson"]
	parsedCredential, err := protocol.ParseCredentialRequestResponseBytes([]byte(responseJSONStr))
	if err != nil {
		return "failure", fmt.Errorf("failed to parse credential assertion: %w", err)
	}

	// Get user from DB
	userModel, err := Services.UserRepo.GetByUsername(ctx, username)
	if err != nil {
		return "failure", fmt.Errorf("user not found: %w", err)
	}

	var storedCredential webauthn.Credential
	if err := json.Unmarshal([]byte(userModel.Attributes["webauthn_credential"]), &storedCredential); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal stored credential: %w", err)
	}

	user := &WebAuthnUserCredentials{
		Username:    username,
		ID:          []byte(username),
		Credentials: []webauthn.Credential{storedCredential},
	}

	// WebAuthn verify
	wconfig := &webauthn.Config{
		RPDisplayName: "Go IAM",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}

	webAuth, err := webauthn.New(wconfig)
	if err != nil {
		return "failure", fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	_, err = webAuth.ValidateLogin(user, session, parsedCredential)
	if err != nil {
		return "failure", fmt.Errorf("passkey login failed: %w", err)
	}

	// Store user in flow context
	userBytes, err := json.Marshal(user)
	if err != nil {
		return "failure", fmt.Errorf("failed to serialize user: %w", err)
	}
	state.Context["user"] = string(userBytes)

	log.Printf("User %s successfully verified via passkey", username)
	return "success", nil
}

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
