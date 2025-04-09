package unit

import (
	"context"
	"testing"

	"goiam/internal/auth/flows"
	"goiam/internal/db/model"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// mockUserRepo implements repository.UserRepository
type mockUserRepo struct {
	user  *model.User
	calls map[string]int
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*model.User, error) {
	m.calls["GetByUsername"]++
	if m.user != nil && m.user.Username == username {
		return m.user, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	m.calls["Create"]++
	return nil
}

func (m *mockUserRepo) Update(_ context.Context, user *model.User) error {
	m.calls["Update"]++
	m.user = user // mimic DB update
	return nil
}

func TestUsernamePasswordFlow_HappyPath(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)

	mockRepo := &mockUserRepo{
		user: &model.User{
			ID:           "user-123",
			Username:     "admin",
			PasswordHash: string(hashed),
		},
		calls: make(map[string]int),
	}

	state := &flows.FlowState{
		RunID: "test-happy",
		Steps: []flows.FlowStep{
			{Name: "init"},
			{Name: "sendUsername", Prompts: map[string]string{
				"username": "admin",
			}},
			{Name: "sendPassword", Prompts: map[string]string{
				"password": "admin",
			}},
		},
	}

	flow := flows.NewUsernamePasswordFlow(mockRepo)
	flow.Run(state)

	assert.Nil(t, state.Error)
	assert.NotNil(t, state.Result)
	assert.Equal(t, "admin", state.Result.Username)
	assert.True(t, state.Result.Authenticated)
	assert.Equal(t, 0, mockRepo.user.FailedLoginAttempts)
}

func TestUsernamePasswordFlow_InvalidPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)

	mockRepo := &mockUserRepo{
		user: &model.User{
			ID:           "user-123",
			Username:     "admin",
			PasswordHash: string(hashed),
		},
		calls: make(map[string]int),
	}

	state := &flows.FlowState{
		RunID: "test-invalid-password",
		Steps: []flows.FlowStep{
			{Name: "init"},
			{Name: "sendUsername", Prompts: map[string]string{
				"username": "admin",
			}},
			{Name: "sendPassword", Prompts: map[string]string{
				"password": "wrongpass",
			}},
		},
	}

	flow := flows.NewUsernamePasswordFlow(mockRepo)
	flow.Run(state)

	assert.Nil(t, state.Result)
	assert.NotNil(t, state.Error)
	assert.Contains(t, *state.Error, "Invalid username or password")
	assert.Greater(t, mockRepo.user.FailedLoginAttempts, 0)
}

func TestUsernamePasswordFlow_TriggersSendUsernameAfterInit(t *testing.T) {
	state := &flows.FlowState{
		RunID: "test-send-username",
		Steps: []flows.FlowStep{{Name: "init"}},
	}

	mockRepo := &mockUserRepo{calls: make(map[string]int)}
	flow := flows.NewUsernamePasswordFlow(mockRepo)
	flow.Run(state)

	assert.Len(t, state.Steps, 2)
	next := state.Steps[1]
	assert.Equal(t, "sendUsername", next.Name)
	assert.Contains(t, next.Prompts, "username")
}

func TestUsernamePasswordFlow_TriggersSendPasswordAfterSendUsername(t *testing.T) {
	state := &flows.FlowState{
		RunID: "test-send-password",
		Steps: []flows.FlowStep{
			{Name: "init"},
			{Name: "sendUsername", Prompts: map[string]string{
				"username": "admin",
			}},
		},
	}

	mockRepo := &mockUserRepo{calls: make(map[string]int)}
	flow := flows.NewUsernamePasswordFlow(mockRepo)
	flow.Run(state)

	assert.Len(t, state.Steps, 3)
	next := state.Steps[2]
	assert.Equal(t, "sendPassword", next.Name)
	assert.Contains(t, next.Prompts, "password")
}
