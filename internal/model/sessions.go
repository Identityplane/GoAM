package model

import "time"

// Represents a ongoing execution of a flow
type AuthenticationSession struct {
	RunID                    string            `json:"run_id"`
	SessionIdHash            string            `json:"session_id_hash"`
	FlowId                   string            `json:"flow_id"`
	Current                  string            `json:"current"` // active node
	Context                  map[string]string `json:"context"` // dynamic values (inputs + outputs)
	History                  []string          `json:"history"` // executed node names
	Error                    *string           `json:"error,omitempty"`
	Result                   *FlowResult       `json:"result,omitempty"`
	User                     *User             `json:"user,omitempty"`
	Prompts                  map[string]string `json:"prompts,omitempty"` // Prompts to be shown to the user, if applicable
	Oauth2SessionInformation *Oauth2Session    `json:"oauth2_request,omitempty"`
	CreatedAt                time.Time         `json:"created_at"`
	ExpiresAt                time.Time         `json:"expires_at"`
	LoginUri                 string            `json:"login_uri"` // Uri of the login flow
}

type Oauth2Session struct {
	AuthorizeRequest *AuthorizeRequest `json:"authorize_request"`
}

// AuthorizeRequest represents the parameters for the authorization request
type AuthorizeRequest struct {
	ClientID            string   `json:"client_id"`
	RedirectURI         string   `json:"redirect_uri"`
	ResponseType        string   `json:"response_type"`
	Scope               []string `json:"scope"`
	State               string   `json:"state"`
	CodeChallenge       string   `json:"code_challenge"`
	CodeChallengeMethod string   `json:"code_challenge_method"`
	Nonce               string   `json:"nonce"`
	Prompt              string   `json:"prompt"`
}
