package model

import (
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/rs/zerolog"
)

const (
	// Name of system nodes
	NODE_FAILURE_RESULT = "failureResult"
	NODE_SUCCESS_RESULT = "successResult"
	NODE_ERROR          = "error"
)

// Represents a ongoing execution of a flow
type AuthenticationSession struct {
	RealmObject
	RunID                        string             `json:"run_id"`            // Unique identifier for the flow execution
	SessionIdHash                string             `json:"session_id_hash"`   // Hash of the session id
	FlowId                       string             `json:"flow_id"`           // Id of the flow
	Current                      string             `json:"current"`           // name of the active node
	CurrentType                  string             `json:"current_type"`      // type of the active node
	Context                      map[string]string  `json:"context"`           // dynamic values (inputs + outputs)
	History                      []string           `json:"history"`           // executed node names
	Error                        *string            `json:"error,omitempty"`   // Error message to be displayed to the user
	Result                       *FlowResult        `json:"result,omitempty"`  // Result of the flow execution
	User                         *User              `json:"user,omitempty"`    // The loaded user from the database
	Prompts                      map[string]string  `json:"prompts,omitempty"` // Prompts to be shown to the user, if applicable
	Oauth2SessionInformation     *Oauth2Session     `json:"oauth2_request,omitempty"`
	SimpleAuthSessionInformation *SimpleAuthContext `json:"simple_auth_request,omitempty"`
	CreatedAt                    time.Time          `json:"created_at"` // Time the session was created
	ExpiresAt                    time.Time          `json:"expires_at"` // Time when this auth session will expire

	// LoginUri is the uri of the login endpoints where the actual authentication takes place
	LoginUriBase string `json:"login_uri_base"`
	LoginUriNext string `json:"login_uri_next"`

	// FinishUri is the uri where the user will be redirected after the flow has been completed
	// When using oaoth2 the graph handler will use this to redirect back oauth2/finishauthorize which will
	// itself redirect to the client application
	FinishUri string `json:"finish_uri"` // Uri of the finish endpoint

	// Debug is a flag to enable debug mode
	// This will add additional debug logs as well as the render to display the debug information which contains sensitive information
	Debug bool `json:"debug"`

	// HttpAuthContext is the context for the http authentication
	HttpAuthContext *HttpAuthContext `json:"http_auth_context,omitempty"`
}

func (s *AuthenticationSession) GetLatestHistory() string {
	if len(s.History) == 0 {
		return ""
	}
	return s.History[len(s.History)-1]
}

func (s *AuthenticationSession) Finished() bool {
	return s.CurrentType == NODE_SUCCESS_RESULT || s.CurrentType == NODE_FAILURE_RESULT || s.CurrentType == NODE_ERROR
}

// DidResultAuthenticated returns true if the result node was a success result and the user is set in the context
func (s *AuthenticationSession) DidResultAuthenticated() bool {

	if s.Result == nil {
		return false
	}

	if s.Result.UserID == "" {
		return false
	}

	if s.CurrentType != NODE_SUCCESS_RESULT {
		return false
	}

	return true
}

// DidResultError returns true if the result node was an error node
func (s *AuthenticationSession) DidResultError() bool {
	return s.CurrentType == NODE_ERROR
}

// DidResultFailure returns true if the result node was a failure result
func (s *AuthenticationSession) DidResultFailure() bool {
	return s.CurrentType == NODE_FAILURE_RESULT
}

// GetLogger returns a zerolog logger with contextual information from the session
func (s *AuthenticationSession) GetLogger() zerolog.Logger {
	log := logger.GetGoamLogger()

	// Add session context to the logger
	event := log.With().
		Str("session_id", s.SessionIdHash[:8]). // First 8 chars for readability
		Str("run_id", s.RunID).
		Str("flow_id", s.FlowId).
		Str("current_node", s.Current)

	// Add user context if available
	if s.User != nil {
		event = event.Str("user_id", s.User.ID)
	}

	// Add OAuth2 context if available
	if s.Oauth2SessionInformation != nil && s.Oauth2SessionInformation.AuthorizeRequest != nil {
		event = event.Str("client_id", s.Oauth2SessionInformation.AuthorizeRequest.ClientID)
	}

	return event.Logger()
}

type Oauth2Session struct {
	AuthorizeRequest *AuthorizeRequest `json:"authorize_request"`
	AuthTime         time.Time         `json:"auth_time"`
	Acr              string            `json:"acr"`
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
	Request             string   `json:"request"`

	// Advanced OIDC parameters
	MaxAge      *int     `json:"max_age"`
	Nonce       string   `json:"nonce"`
	Prompt      string   `json:"prompt"`
	IdTokenHint string   `json:"id_token_hint"`
	AcrValues   []string `json:"acr_values"`
}
