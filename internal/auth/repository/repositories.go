package repository

import (
	"context"

	"github.com/gianlucafrei/GoAM/internal/model"
)

type Repositories struct {
	UserRepo    UserRepository
	EmailSender EmailSender
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByLoginIdentifier(ctx context.Context, loginIdentifier string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByFederatedIdentifier(ctx context.Context, provider, identifier string) (*model.User, error)

	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}

type EmailSender interface {
	SendEmail(subject, body, recipientEmail, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail string) error
}
