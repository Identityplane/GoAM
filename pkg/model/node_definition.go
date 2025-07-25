package model

import (
	"context"
)

// NodeDefinition is a definition of a node in the graph
type NodeDefinition struct {
	Name                 string            // e.g. "askUsername", references as use
	PrettyName           string            // "Ask Username"
	Description          string            // Description of the node as text
	Category             string            // Category for the editor
	Type                 NodeType          // query, logic, etc.
	RequiredContext      []string          `json:"inputs"`  // field that the node requires from the flow context
	OutputContext        []string          `json:"outputs"` // fields that the node will set in the flow context
	PossiblePrompts      map[string]string `json:"prompts"` // key: label/type shown to user, will be returned via the user input argument
	PossibleResultStates []string
	CustomConfigOptions  map[string]string                                                                                                         // e.g. ["success", "fail"]
	Run                  func(state *AuthenticationSession, node *GraphNode, input map[string]string, services *Repositories) (*NodeResult, error) // Run function for logic nodes, must either return a condition or a set of prompts
}

type Repositories struct {
	UserRepo    UserRepository
	EmailSender EmailSender
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByLoginIdentifier(ctx context.Context, loginIdentifier string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByFederatedIdentifier(ctx context.Context, provider, identifier string) (*User, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

type EmailSender interface {
	SendEmail(subject, body, recipientEmail, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail string) error
}
