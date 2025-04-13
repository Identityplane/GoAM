package graph

import (
	"context"
	"fmt"
	"goiam/internal/db/model"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var CreateUserNode = &NodeDefinition{
	Name:            "createUser",
	Type:            NodeTypeLogic,
	RequiredContext: []string{"username", "password"},
	OutputContext:   []string{"user_id"},
	Conditions:      []string{"success", "fail"},
	Run:             RunCreateUserNode,
}

func RunCreateUserNode(state *FlowState, node *GraphNode, input map[string]string) (*NodeResult, error) {
	ctx := context.Background()
	username := state.Context["username"]
	password := state.Context["password"]

	userRepo := Services.UserRepo
	if userRepo == nil {
		return NewNodeResultWithTextError("UserRepo not initialized")
	}

	// Check for existing user
	existing, _ := userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return NewNodeResultWithTextError("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return NewNodeResultWithError(fmt.Errorf("failed to hash password: %w", err))
	}

	user := &model.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return NewNodeResultWithError(fmt.Errorf("failed to create user: %w", err))
	}

	state.User = user
	state.Context["user_id"] = user.ID
	state.Context["username"] = user.Username

	state.Result = &FlowResult{
		UserID:        user.ID,
		Username:      user.Username,
		Authenticated: true,
		AuthLevel:     AuthLevel1FA,
		FlowName:      "user_register",
	}

	return NewNodeResultWithCondition("success")
}
