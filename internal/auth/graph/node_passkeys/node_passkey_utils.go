package node_passkeys

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/go-webauthn/webauthn/webauthn"
)

func credIdToString(credId []byte) string {
	return base64.StdEncoding.EncodeToString(credId)
}

func generatePasskeysOptions(state *model.AuthenticationSession, node *model.GraphNode, accountName string, userId string) (map[string]string, error) {

	optionsJSON, err := GeneratePasskeysChallenge(state, node, accountName, userId)

	if err != nil {
		return nil, fmt.Errorf("failed to generate passkeys challenge: %w", err)
	}

	prompts := &map[string]string{"passkeysOptions": string(optionsJSON)}

	return *prompts, nil
}

func GeneratePasskeysChallenge(state *model.AuthenticationSession, node *model.GraphNode, accountName string, userId string) (string, error) {

	// Get WebAuthN config
	wconfig, err := getWebAuthnConfig(state, node)
	if err != nil {
		return "", fmt.Errorf("failed to get webauthn config: %w", err)
	}
	webAuth, err := webauthn.New(wconfig)

	if err != nil {
		return "", fmt.Errorf("failed to initialize webauthn: %w", err)
	}

	// Create user credentials object
	userCredentials := &WebAuthnUserCredentials{
		Username:    accountName,
		DisplayName: accountName,
		ID:          []byte(userId),
	}
	options, session, err := webAuth.BeginRegistration(userCredentials)

	if err != nil {
		return "", fmt.Errorf("failed to start passkeys registartion %w", err)
	}

	// Marshal session which is the server side session data of the passkey challenge
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	// Marshal options which is the client side options of the passkey challenge
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("failed to marshal options: %w", err)
	}

	state.Context["passkeysSession"] = string(sessionJSON)
	state.Context["passkeysOptions"] = string(optionsJSON)
	return string(optionsJSON), nil
}

func getWebAuthnConfig(state *model.AuthenticationSession, node *model.GraphNode) (*webauthn.Config, error) {

	// Passkeys need subdomain in oder to be properly separated between different tenants
	/* Relying Party Identifier https://www.w3.org/TR/webauthn-2/
	RP ID
	In the context of the WebAuthn API, a relying party identifier is a valid domain string identifying the WebAuthn Relying Party on whose behalf a given registration or authentication ceremony is being performed. A public key credential can only be used for authentication with the same entity (as identified by RP ID) it was registered with.

	By default, the RP ID for a WebAuthn operation is set to the caller's origin's effective domain. This default MAY be overridden by the caller, as long as the caller-specified RP ID value is a registrable domain suffix of or is equal to the caller's origin's effective domain. See also § 5.1.3 Create a New Credential - PublicKeyCredential's [[Create]](origin, options, sameOriginWithAncestors) Method and § 5.1.4 Use an Existing Credential to Make an Assertion - PublicKeyCredential's [[Get]](options) Method.*/

	// Get the RPOrigins for the login uri which is everything without the path
	// Step 1 parse the login uri with net/url
	loginUri, err := url.Parse(state.LoginUriBase)

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
