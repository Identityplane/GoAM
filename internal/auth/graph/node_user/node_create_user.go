package node_user

import (
	"context"
	"fmt"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

var CreateUserNode = &model.NodeDefinition{
	Name:            "createUser",
	PrettyName:      "Create User",
	Description:     "Creates a new user account in the database with the provided username and password",
	Category:        "User Management",
	Type:            model.NodeTypeLogic,
	RequiredContext: []string{"username", "password", "email"},
	CustomConfigOptions: map[string]string{
		"checkUsernameUnique": "If set to 'true' the username will be checked for uniqueness. In that case username must be present in the context or user object.",
		"checkEmailUnique":    "If set to 'true' the email will be checked for uniqueness. In that case email must be present in the context or user object.",
		"skipSaveUser":        "If set to 'true' the user will not be saved to the database after creation and only the context will be updated",
	},
	OutputContext:        []string{"user_id"},
	PossibleResultStates: []string{"success", "existing"},
	Run:                  RunCreateUserNode,
}

func RunCreateUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {
	ctx := context.Background()

	// Check if the username is unique
	if node.CustomConfig["checkUsernameUnique"] == "true" {
		existing, err := services.UserRepo.GetByAttributeIndex(ctx, model.AttributeTypeUsername, state.Context["username"])
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to get user by username: %w", err))
		}
		if existing != nil {
			return model.NewNodeResultWithCondition("existing")
		}
	}

	// Check if the email is unique
	if node.CustomConfig["checkEmailUnique"] == "true" {
		existing, err := services.UserRepo.GetByAttributeIndex(ctx, model.AttributeTypeEmail, state.Context["email"])
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to get user by email: %w", err))
		}
		if existing != nil {
			return model.NewNodeResultWithCondition("existing")
		}
	}

	// Create a user if we dont already have one in the context
	if state.User == nil {
		var err error
		state.User, err = services.UserRepo.NewUserModel(state)

		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to create user: %w", err))
		}
	}

	// Check if the user already exists in the database
	existing, err := services.UserRepo.GetByID(ctx, state.User.ID)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}
	if existing != nil {
		return model.NewNodeResultWithCondition("existing")
	}

	// If we have a username in the context but no username attribute we create a username attribute
	username := state.Context["username"]
	existingUsernames, _, err := model.GetAttributes[model.UsernameAttributeValue](state.User, model.AttributeTypeUsername)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to get username attributes: %w", err))
	}
	// Check if any of the existing usernames are the same as the username in the context
	alreadyExists := false
	for _, existingUsername := range existingUsernames {
		if existingUsername.PreferredUsername == username {
			alreadyExists = true
		}
	}
	if username != "" && !alreadyExists {
		usernameAttribute := &model.UserAttribute{
			Type:  model.AttributeTypeUsername,
			Index: &username,
			Value: model.UsernameAttributeValue{
				PreferredUsername: username,
			},
		}
		state.User.AddAttribute(usernameAttribute)
	}

	// If we have an email in the context we create an email attribute
	email := state.Context["email"]
	existingEmails, _, err := model.GetAttributes[model.EmailAttributeValue](state.User, model.AttributeTypeEmail)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to get email attributes: %w", err))
	}
	// Check if any of the existing emails are the same as the email in the context
	alreadyExists = false
	for _, existingEmail := range existingEmails {
		if existingEmail.Email == email {
			alreadyExists = true
		}
	}
	if email != "" && !alreadyExists {

		emailAttribute := &model.UserAttribute{
			Type:  model.AttributeTypeEmail,
			Index: &email,
			Value: model.EmailAttributeValue{
				Email:    email,
				Verified: false,
			},
		}
		state.User.AddAttribute(emailAttribute)
	}

	// If we have a password in the context we create a password attribute
	// TODO Should we check that we dont already have a password attribute?
	password := state.Context["password"]
	if password != "" {
		hashed, err := lib.HashPassword(password)
		if err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to hash password: %w", err))
		}

		passwordAttribute := &model.UserAttribute{
			Type: model.AttributeTypePassword,
			Value: model.PasswordAttributeValue{
				PasswordHash:   string(hashed),
				Locked:         false,
				FailedAttempts: 0,
			},
		}
		state.User.AddAttribute(passwordAttribute)
	}

	if node.CustomConfig["skipSaveUser"] != "true" {
		if err := services.UserRepo.Create(ctx, state.User); err != nil {
			return model.NewNodeResultWithError(fmt.Errorf("failed to create user: %w", err))
		}
	}

	return model.NewNodeResultWithCondition("success")
}
