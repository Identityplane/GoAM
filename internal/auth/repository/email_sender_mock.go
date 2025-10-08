package repository

import (
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/stretchr/testify/mock"
)

// MockEmailSender implements EmailSender for testing
type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendEmail(email *model.SendEmailParams) error {
	args := m.Called(email)
	return args.Error(0)
}

// NewMockEmailSender creates a new mock email sender
func NewMockEmailSender() *MockEmailSender {
	return new(MockEmailSender)
}
