package attributes

import (
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/security"
	"github.com/google/uuid"
)

type ActionToken struct {
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	Action    string    `json:"action"`
}

func (a *ActionToken) GetIndex() string {
	return a.TokenHash
}

func (a *ActionToken) IndexIsSensitive() bool {
	return true
}

// CreateFlowToken creates a new action token
func CreateFlowToken(action string, expiresIn time.Duration) (*ActionToken, string, error) {
	secret := uuid.New().String()
	tokenHash := security.HashString(secret)
	return &ActionToken{
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(expiresIn),
		Action:    action,
	}, secret, nil
}

// ValidateActionToken validates an action token
// checks if it is not expired and if the action matches
func (a *ActionToken) ValidateActionToken(token string, action string) error {

	// If the hash does not match, return an error
	tokenHash := security.HashString(token)
	if tokenHash != a.TokenHash {
		return fmt.Errorf("token hash mismatch")
	}

	// if the action does not match, return an error
	if action != a.Action {
		return fmt.Errorf("action mismatch")
	}

	// if the token is expired, return an error
	if a.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	return nil
}
