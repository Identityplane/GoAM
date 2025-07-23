package model

import (
	"encoding/json"
	"time"
)

// PersistentAuthSession represents the database model for persisting AuthenticationSession
type PersistentAuthSession struct {
	Tenant             string    `json:"tenant"`
	Realm              string    `json:"realm"`
	RunID              string    `json:"run_id"`
	SessionIDHash      string    `json:"session_id_hash"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	SessionInformation []byte    `json:"session_information"`
}

// NewPersistentAuthSession creates a new PersistentAuthSession from an AuthenticationSession
func NewPersistentAuthSession(tenant, realm string, session *AuthenticationSession) (*PersistentAuthSession, error) {
	sessionInfo, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	return &PersistentAuthSession{
		Tenant:             tenant,
		Realm:              realm,
		RunID:              session.RunID,
		SessionIDHash:      session.SessionIdHash,
		CreatedAt:          session.CreatedAt,
		ExpiresAt:          session.ExpiresAt,
		SessionInformation: sessionInfo,
	}, nil
}

// ToAuthenticationSession converts the PersistentAuthSession back to an AuthenticationSession
func (s *PersistentAuthSession) ToAuthenticationSession() (*AuthenticationSession, error) {
	var session AuthenticationSession
	err := json.Unmarshal(s.SessionInformation, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
