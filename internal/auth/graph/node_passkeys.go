package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/model"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

var PasskeyRegisterNode = &NodeDefinition{
	Name:            "registerPasskey",
	PrettyName:      "Register Passkey",
	Description:     "Registers a new passkey (WebAuthn credential) for the user to enable passwordless authentication",
	Category:        "Passkeys",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{"username"},
	PossiblePrompts: map[string]string{
		"passkeysFinishRegistrationJson": "json",
	},
	OutputContext:        []string{"passkeysFinishRegistrationJson", "passkeysSession", "passkeysOptions"},
	PossibleResultStates: []string{"success", "failure"},
	Run:                  RunPasskeyRegisterNode,
}

var PasskeysVerifyNode = &NodeDefinition{
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

var PasskeysCheckUserRegistered = &NodeDefinition{
	Name:                 "checkPasskeyRegistered",
	PrettyName:           "Check Passkey Registration",
	Description:          "Checks if a user has already registered a passkey for passwordless authentication",
	Category:             "Passkeys",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"username"},
	PossiblePrompts:      nil,
	OutputContext:        []string{},
	PossibleResultStates: []string{"registered", "not_registered", "user_not_found"},
	Run:                  RunCheckUserHasPasskeyNode,
}

var AskEnrollPasskeyNode = &NodeDefinition{
	Name:                 "askEnrollPasskey",
	PrettyName:           "Ask to Enroll Passkey",
	Description:          "Prompts the user to choose whether they want to enroll a passkey for future passwordless authentication",
	Category:             "Passkeys",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{},
	PossiblePrompts:      map[string]string{"enrollPasskey": "boolean"},
	OutputContext:        []string{"enrollPasskey"},
	PossibleResultStates: []string{"yes", "no"},
	Run:                  RunAskEnrollPasskeyNode,
}

// Very simple node that asks the user if they want to enroll a passkey
func RunAskEnrollPasskeyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	enrollPasskey := input["enrollPasskey"]

	if enrollPasskey == "" {
		return model.NewNodeResultWithPrompts(map[string]string{"enrollPasskey": "boolean"})
	} else if enrollPasskey == "true" {
		return model.NewNodeResultWithCondition("yes")
	} else {
		return model.NewNodeResultWithCondition("no")
	}

}

func tryLoadUserFromFlowContext(state *model.AuthenticationSession, services *repository.Repositories) (*model.User, error) {

	userId := state.Context["user_id"]
	username := state.Context["username"]
	email := state.Context["email"]
	loginIdentifier := state.Context["loginIdentifier"]

	if userId != "" {
		user, err := services.UserRepo.GetByID(context.Background(), userId)

		if err != nil || user != nil {
			return user, err
		}
	}
	if username != "" {
		user, err := services.UserRepo.GetByUsername(context.Background(), username)

		if err != nil || user != nil {
			return user, err
		}
	}
	if email != "" {
		user, err := services.UserRepo.GetByEmail(context.Background(), email)

		if err != nil || user != nil {
			return user, err
		}
	}
	if loginIdentifier != "" {
		user, err := services.UserRepo.GetByLoginIdentifier(context.Background(), loginIdentifier)

		if err != nil || user != nil {
			return user, err
		}
	}

	return nil, nil
}

func RunCheckUserHasPasskeyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	// Check if user is already loaded, if not we load it by id/email/username
	user := state.User
	var err error
	if user == nil {
		user, err = tryLoadUserFromFlowContext(state, services)

		if err != nil {
			return model.NewNodeResultWithError(err)
		}
	}

	if user == nil {
		return model.NewNodeResultWithCondition("user_not_found")
	}

	// Check if the user has a passkey registered
	if user.WebAuthnCredential == "" {
		state.Context["hasPasskeyRegistered"] = "not_registered"
		return model.NewNodeResultWithCondition("not_registered")
	} else {
		state.Context["hasPasskeyRegistered"] = "registered"
		return model.NewNodeResultWithCondition("registered")
	}
}

func RunPasskeyRegisterNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	// Check if input is present, if not generate options, if present process registration
	if _, ok := input["passkeysFinishRegistrationJson"]; !ok {

		// Generate options
		prompts, err := GeneratePasskeysOptions(state, node)
		if err != nil {

			return model.NewNodeResultWithError(fmt.Errorf("failed to generate passkeys options: %w", err))
		}
		return model.NewNodeResultWithPrompts(prompts)

	} else {
		// Process registration
		result, err := ProcessPasskeyRegistration(state, node, input, services)
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to process passkey registration: %w", err))
		}
		return model.NewNodeResultWithCondition(result)
	}
}

func RunPasskeyVerifyNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

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
	passkeysLoginOptions, err := generatePasskeysChallenge(state, node, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate passkey challenge: %w", err)
	}

	return model.NewNodeResultWithPrompts(map[string]string{"passkeysLoginOptions": passkeysLoginOptions})
}

func getWebAuthnConfig(state *model.AuthenticationSession, node *model.GraphNode) (*webauthn.Config, error) {

	// Passkeys need subdomain in oder to be properly separated between different tenants
	/* Relying Party Identifier https://www.w3.org/TR/webauthn-2/
	RP ID
	In the context of the WebAuthn API, a relying party identifier is a valid domain string identifying the WebAuthn Relying Party on whose behalf a given registration or authentication ceremony is being performed. A public key credential can only be used for authentication with the same entity (as identified by RP ID) it was registered with.

	By default, the RP ID for a WebAuthn operation is set to the caller's origin's effective domain. This default MAY be overridden by the caller, as long as the caller-specified RP ID value is a registrable domain suffix of or is equal to the caller's origin's effective domain. See also § 5.1.3 Create a New Credential - PublicKeyCredential's [[Create]](origin, options, sameOriginWithAncestors) Method and § 5.1.4 Use an Existing Credential to Make an Assertion - PublicKeyCredential's [[Get]](options) Method.*/

	// Get the RPOrigins for the login uri which is everything without the path
	// Step 1 parse the login uri with net/url
	loginUri, err := url.Parse(state.LoginUri)

	if err != nil {
		return nil, fmt.Errorf("failed to parse login uri: %w", err)
	}

	// Step 2 get the hostname

	// The RpId is the hostname of the RPOrigins without the procotol https://manage.identityplane.cloud -> manage.identityplane.cloud
	// If we have a custom rpId in the node config, we use it, otherwise we use the origin
	rpId := loginUri.Hostname()
	if node != nil && node.CustomConfig != nil && node.CustomConfig["rpId"] != "" {
		rpId = node.CustomConfig["rpId"]
	}

	// the RPOrigin is the protocol + hostname + port (if present) unless it is also overwritten in the node config
	protocol := loginUri.Scheme
	host := loginUri.Host // This includes both hostname and port if present

	// If we have a custom rpId, we need to construct the origin with the custom rpId but keep the port from the original host
	var rpOrigin string
	if node != nil && node.CustomConfig != nil && node.CustomConfig["rpOrigin"] != "" {
		rpOrigin = node.CustomConfig["rpOrigin"]
	} else if node != nil && node.CustomConfig != nil && node.CustomConfig["rpId"] != "" {
		// Use custom rpId but keep the port from original host
		if strings.Contains(host, ":") {
			port := strings.Split(host, ":")[1]
			rpOrigin = protocol + "://" + rpId + ":" + port
		} else {
			rpOrigin = protocol + "://" + rpId
		}
	} else {
		rpOrigin = protocol + "://" + host
	}

	// If we have a custom rpDisplayName in the node config, we use it, otherwise we use "Go IAM"
	rpDisplayName := "Go IAM"
	if node != nil && node.CustomConfig != nil && node.CustomConfig["rpDisplayName"] != "" {
		rpDisplayName = node.CustomConfig["rpDisplayName"]
	}

	return &webauthn.Config{
		RPDisplayName: rpDisplayName,
		RPID:          rpId,
		RPOrigins:     []string{rpOrigin},
	}, nil
}

func GeneratePasskeysOptions(state *model.AuthenticationSession, node *model.GraphNode) (map[string]string, error) {
	//ctx := context.Background()
	user := state.User
	if user == nil {
		return nil, fmt.Errorf("user must be loaded before registering a passkey")
	}

	optionsJSON, err := generatePasskeysChallenge(state, node, user.Username, user.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to generate passkeys challenge: %w", err)
	}

	prompts := &map[string]string{"passkeysOptions": string(optionsJSON)}

	return *prompts, nil
}

func generatePasskeysChallenge(state *model.AuthenticationSession, node *model.GraphNode, username string, userId string) (string, error) {

	// Init passkys
	wconfig, err := getWebAuthnConfig(state, node)
	if err != nil {
		return "", fmt.Errorf("failed to get webauthn config: %w", err)
	}
	webAuth, err := webauthn.New(wconfig)

	if err != nil {
		return "", fmt.Errorf("failed to initialize webauthn: %w", err)
	}

	userCredentials := &WebAuthnUserCredentials{
		Username: username,
		ID:       []byte(userId),
	}
	options, session, err := webAuth.BeginRegistration(userCredentials)

	if err != nil {
		return "", fmt.Errorf("failed to start passkeys registartion %w", err)
	}

	// Marshal session and options into JSON strings
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("failed to marshal options: %w", err)
	}

	state.Context["passkeysSession"] = string(sessionJSON)
	state.Context["passkeysOptions"] = string(optionsJSON)

	return string(optionsJSON), nil
}

func ProcessPasskeyRegistration(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (string, error) {

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
	user := state.User
	userCredentials := &WebAuthnUserCredentials{
		Username: user.Username,
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

	// Store new credential with user
	userRepo := services.UserRepo
	if userRepo == nil {
		return "fail", errors.New("userRepo not initialized")
	}

	credBytes, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}

	user.LoginIdentifier = string(parsedCredential.ID)
	user.WebAuthnCredential = string(credBytes)

	// Then safe the user with updated credential
	err = userRepo.Update(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to update user: %w", err)
	}

	log := logger.GetLogger()
	log.Info().Str("credential_id", string(cred.ID)).Msg("successfully registered credential")
	return "success", nil
}

func GeneratePasskeysLoginOptions(state *model.AuthenticationSession, node *model.GraphNode, services *repository.Repositories) (map[string]string, error) {

	// Setup config
	wconfig, err := getWebAuthnConfig(state, node)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webauthn config: %w", err)
	}
	webAuth, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webauthn config: %w", err)
	}

	userCredentials := &WebAuthnUserCredentials{
		Username: "",
		ID:       []byte(""),
	}

	options, session, err := webAuth.BeginLogin(userCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to start passkey login: %w", err)
	}

	// Store in context
	sessionJSON, _ := json.Marshal(session)
	optionsJSON, _ := json.Marshal(options)
	state.Context["passkeysSession"] = string(sessionJSON)
	state.Context["passkeysLoginOptions"] = string(optionsJSON)

	return map[string]string{
		"passkeysLoginOptions": string(optionsJSON),
	}, nil
}

func ProcessPasskeyLogin(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (string, error) {
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

	// Load user by login identifier
	user, err := services.UserRepo.GetByLoginIdentifier(context.Background(), credentialID)
	if err != nil {
		return "failure", fmt.Errorf("failed to load user by login identifier: %w", err)
	}
	// Todo, if not present we need to load the user from the database

	if user == nil {
		return "failure", fmt.Errorf("user not found")
	}

	// Copy over from custom attributes if it was not set correctly
	if user.WebAuthnCredential == "" && user.Attributes["webauthn_credential"] != "" {
		user.WebAuthnCredential = user.Attributes["webauthn_credential"]
		services.UserRepo.Update(context.Background(), user)
	}

	// if we still have no webauthn credential, we return failure
	if user.WebAuthnCredential == "" {
		return "failure", fmt.Errorf("user has no registered passkey")
	}

	var storedCredential webauthn.Credential
	if err := json.Unmarshal([]byte(user.WebAuthnCredential), &storedCredential); err != nil {
		return "failure", fmt.Errorf("failed to unmarshal stored credential: %w", err)
	}

	userCredentials := &WebAuthnUserCredentials{
		Username:    user.Username,
		ID:          []byte(user.ID),
		Credentials: []webauthn.Credential{storedCredential},
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
		session.UserID = []byte(user.ID)
	}

	_, err = webAuth.ValidateLogin(userCredentials, session, parsedCredential)
	if err != nil {
		return "failure", fmt.Errorf("passkey login failed: %w", err)
	}

	state.User = user

	log.Debug().Str("user_id", user.ID).Msg("user successfully verified via passkey")
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
