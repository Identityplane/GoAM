package graph

import (
	"context"
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/model"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var CreateUserNode = &NodeDefinition{
	Name:            "createUser",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{"username", "password"},
	OutputContext:   []string{"user_id"},
	Conditions:      []string{"success", "fail"},
	Run:             RunCreateUserNode,
}

func RunCreateUserNode(state *model.FlowState, node *model.GraphNode, input map[string]string, services *repository.ServiceRegistry) (*model.NodeResult, error) {
	ctx := context.Background()
	username := state.Context["username"]
	password := state.Context["password"]

	userRepo := services.UserRepo
	if userRepo == nil {
		return model.NewNodeResultWithTextError("UserRepo not initialized")
	}

	// Check for existing user
	existing, _ := userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return model.NewNodeResultWithTextError("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to hash password: %w", err))
	}

	user := &model.User{
		ID:                 uuid.NewString(),
		Username:           username,
		PasswordCredential: string(hashed),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to create user: %w", err))
	}

	state.User = user
	state.Context["user_id"] = user.ID
	state.Context["username"] = user.Username

	state.Result = &model.FlowResult{
		UserID:        user.ID,
		Username:      user.Username,
		Authenticated: true,
		AuthLevel:     model.AuthLevel1FA,
		FlowName:      "user_register",
	}

	return model.NewNodeResultWithCondition("success")
}
