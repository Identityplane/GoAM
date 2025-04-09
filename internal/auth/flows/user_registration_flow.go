package flows

import (
	"context"
	"time"

	"goiam/internal/auth/repository"
	"goiam/internal/db/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRegistrationFlow struct {
	UserRepo repository.UserRepository
}

func NewUserRegistrationFlow(repo repository.UserRepository) *UserRegistrationFlow {
	return &UserRegistrationFlow{UserRepo: repo}
}

func (f *UserRegistrationFlow) Run(state *FlowState) {
	last := LastStep(state)
	if last == nil || last.Name == "" {
		state.Steps = append(state.Steps, FlowStep{Name: "init"})
		return
	}

	switch last.Name {
	case "init":
		f.askUsername(state)
	case "askUsername":
		f.askPassword(state)
	case "askPassword":
		f.createUser(state)
	default:
		msg := "Invalid step: " + last.Name
		state.Error = &msg
	}
}

func (f *UserRegistrationFlow) askUsername(state *FlowState) {
	state.Steps = append(state.Steps, FlowStep{
		Name: "askUsername",
		Parameters: map[string]string{
			"username": "",
		},
	})
}

func (f *UserRegistrationFlow) askPassword(state *FlowState) {
	state.Steps = append(state.Steps, FlowStep{
		Name: "askPassword",
		Parameters: map[string]string{
			"password": "",
		},
	})
}

func (f *UserRegistrationFlow) createUser(state *FlowState) {
	ctx := context.Background()

	username := GetStep(state, "askUsername").Parameters["username"]
	password := GetStep(state, "askPassword").Parameters["password"]

	existing, _ := f.UserRepo.GetByUsername(ctx, username)
	if existing != nil {
		msg := "Username already exists"
		state.Error = &msg
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		msg := "Failed to hash password"
		state.Error = &msg
		return
	}

	user := &model.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := f.UserRepo.Create(ctx, user); err != nil {
		msg := "Failed to create user: " + err.Error()
		state.Error = &msg
		return
	}

	state.Result = &FlowResult{
		UserID:        user.ID,
		Username:      user.Username,
		Authenticated: true,
		AuthLevel:     AuthLevel1FA,
	}
}
