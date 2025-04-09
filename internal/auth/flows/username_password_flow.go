package flows

import (
	"context"
	"time"

	"goiam/internal/auth/repository"

	"golang.org/x/crypto/bcrypt"
)

type UsernamePasswordFlow struct {
	UserRepo repository.UserRepository
}

func NewUsernamePasswordFlow(repo repository.UserRepository) *UsernamePasswordFlow {
	return &UsernamePasswordFlow{
		UserRepo: repo,
	}
}

func (f *UsernamePasswordFlow) Run(state *FlowState) {
	lastStep := LastStep(state)
	if lastStep == nil {
		state.Steps = append(state.Steps, FlowStep{Name: "init"})
		return
	}

	switch lastStep.Name {
	case "init":
		f.requestUsername(state)
	case "sendUsername":
		f.requestPassword(state)
	case "sendPassword":
		f.verifyLogin(state)
	default:
		msg := "Unexpected step: " + lastStep.Name
		state.Error = &msg
	}
}

func (f *UsernamePasswordFlow) requestUsername(state *FlowState) {
	state.Steps = append(state.Steps, FlowStep{
		Name: "sendUsername",
		Parameters: map[string]string{
			"username": "",
		},
	})
}

func (f *UsernamePasswordFlow) requestPassword(state *FlowState) {
	state.Steps = append(state.Steps, FlowStep{
		Name: "sendPassword",
		Parameters: map[string]string{
			"password": "",
		},
	})
}

func (f *UsernamePasswordFlow) verifyLogin(state *FlowState) {
	ctx := context.Background()

	username := GetStep(state, "sendUsername").Parameters["username"]
	password := GetStep(state, "sendPassword").Parameters["password"]

	user, err := f.UserRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		// Don't reveal user existence
		msg := "Invalid username or password"
		state.Error = &msg
		return
	}

	// Check for lockout
	if user.FailedLoginAttempts >= 3 || user.AccountLocked {
		msg := "User is blocked due to too many failed login attempts"
		state.Error = &msg
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		user.FailedLoginAttempts++
		user.LastFailedLoginAt = ptr(time.Now())

		// Optionally: auto-lock
		if user.FailedLoginAttempts >= 3 {
			user.AccountLocked = true
		}

		_ = f.UserRepo.Update(ctx, user)

		msg := "Invalid username or password"
		state.Error = &msg
		return
	}

	// Success: reset counters
	user.FailedLoginAttempts = 0
	user.AccountLocked = false
	user.LastFailedLoginAt = nil
	user.LastLoginAt = ptr(time.Now())

	_ = f.UserRepo.Update(ctx, user)

	state.Result = &FlowResult{
		UserID:        user.ID,
		Username:      user.Username,
		Authenticated: true,
		AuthLevel:     AuthLevel1FA,
	}
}

func ptr[T any](v T) *T {
	return &v
}
