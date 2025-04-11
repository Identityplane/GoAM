package graph

import (
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
)

var PasskeyRegisterNode = &NodeDefinition{
	Name:              "registerPasskey",
	Type:              NodeTypeQueryWithLogic,
	Inputs:            []string{"username"},
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
		RPDisplayName: "Go IAM",                           // Display Name for your site
		RPID:          "localhost:8080",                   // Generally the FQDN for your site
		RPOrigins:     []string{"http://localhost:8080/"}, // The origin URLs allowed for WebAuthn requests
	}

	webAuth, err := webauthn.New(wconfig)

	if err != nil {
		panic(err)
	}

	user := &WebAuthnUserCredentials{
		Username: username,
	}
	options, session, err := webAuth.BeginRegistration(user)

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

func ProcessPasskeyRegistration(state *FlowState, node *GraphNode) (string, error) {

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
