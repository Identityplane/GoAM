package model

import (
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/rs/zerolog"
)

// Represents a ongoing execution of a flow
type AuthenticationSession struct {
	RunID                    string            `json:"run_id"`            // Unique identifier for the flow execution
	SessionIdHash            string            `json:"session_id_hash"`   // Hash of the session id
	FlowId                   string            `json:"flow_id"`           // Id of the flow
	Current                  string            `json:"current"`           // active node
	Context                  map[string]string `json:"context"`           // dynamic values (inputs + outputs)
	History                  []string          `json:"history"`           // executed node names
	Error                    *string           `json:"error,omitempty"`   // Error message to be displayed to the user
	Result                   *FlowResult       `json:"result,omitempty"`  // Result of the flow execution
	User                     *User             `json:"user,omitempty"`    // The loaded user from the database
	Prompts                  map[string]string `json:"prompts,omitempty"` // Prompts to be shown to the user, if applicable
	Oauth2SessionInformation *Oauth2Session    `json:"oauth2_request,omitempty"`
	CreatedAt                time.Time         `json:"created_at"` // Time the session was created
	ExpiresAt                time.Time         `json:"expires_at"` // Time when this auth session will expire

	// LoginUri is the uri of the login endpoints where the actual authentication takes place
	LoginUri string `json:"login_uri"`

	// FinishUri is the uri where the user will be redirected after the flow has been completed
	// When using oaoth2 the graph handler will use this to redirect back oauth2/finishauthorize which will
	// itself redirect to the client application
	FinishUri string `json:"finish_uri"` // Uri of the finish endpoint

	// Debug is a flag to enable debug mode
	// This will add additional debug logs as well as the render to display the debug information which contains sensitive information
	Debug bool `json:"debug"`
}

func (s *AuthenticationSession) GetLatestHistory() string {
	if len(s.History) == 0 {
		return ""
	}
	return s.History[len(s.History)-1]
}

// GetLogger returns a zerolog logger with contextual information from the session
func (s *AuthenticationSession) GetLogger() zerolog.Logger {
	log := logger.GetLogger()

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
