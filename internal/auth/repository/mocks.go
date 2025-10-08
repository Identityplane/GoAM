package repository

import "github.com/Identityplane/GoAM/pkg/model"

// NewMockRepositories creates a new Repositories struct with mock implementations
func NewMockRepositories() *model.Repositories {
	return &model.Repositories{
		UserRepo:    new(MockUserRepository),
		EmailSender: new(MockEmailSender),
	}
}

// NewTestRepositories creates a new Repositories struct with real SQLite implementations for testing
func NewTestRepositories(tenant, realm string) (*model.Repositories, error) {
	userRepo, err := NewTestUserRepository(tenant, realm)
	if err != nil {
		return nil, err
	}

	return &model.Repositories{
		UserRepo:    userRepo,
		EmailSender: new(MockEmailSender),
	}, nil
}
