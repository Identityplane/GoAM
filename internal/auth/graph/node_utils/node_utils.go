package node_utils

import (
	"context"
	"errors"

	"github.com/Identityplane/GoAM/pkg/model"
)

var (
	ErrorNoUserIdentifierInContext = errors.New("no user identifier in context")
)

// GetAccountNameFromContext returns the display name of the user from the context
// This name is ised in various context where we need to display the id to the user
func GetAccountNameFromContext(state *model.AuthenticationSession) string {

	// If there us no user in the context we return an empty string
	if state.User == nil {
		return ""
	}

	// If we have a username we return it
	usernames, _, err := model.GetAttributes[model.UsernameAttributeValue](state.User, model.AttributeTypeUsername)
	if err == nil && len(usernames) > 0 {
		return usernames[0].Username
	}

	// Find the first verified email otherwise we return the first email
	emails, _, err := model.GetAttributes[model.EmailAttributeValue](state.User, model.AttributeTypeEmail)
	if err == nil && len(emails) > 0 {
		for _, email := range emails {
			if email.Verified {
				return email.Email
			}
		}
		return emails[0].Email
	}

	// Otherwise we return the id
	return state.User.ID
}

// LoadUserFromContext loads the user from the context, returns an error if no user is found
func LoadUserFromContext(state *model.AuthenticationSession, services *model.Repositories) (*model.User, error) {

	// If the user is already set we return it
	if state.User != nil {
		return state.User, nil
	}

	if services == nil || services.UserRepo == nil {
		return nil, errors.New("user repository is not set")
	}

	// If we have a username in the context we load the user from the database
	if state.Context["username"] != "" {
		user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeUsername, state.Context["username"])
		if err != nil {
			return nil, err
		}

		state.User = user
		return user, nil
	}

	// If we have an email in the context we load the user from the database
	if state.Context["email"] != "" {
		user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeEmail, state.Context["email"])
		if err != nil {
			return nil, err
		}

		state.User = user
		return user, nil
	}

	// If we have a phone in the context we load the user from the database
	if state.Context["phone"] != "" {
		user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypePhone, state.Context["phone"])
		if err != nil {
			return nil, err
		}

		state.User = user
		return user, nil
	}

	// If we have a user_id in the context we load the user from the database
	if state.Context["user_id"] != "" {
		user, err := services.UserRepo.GetByID(context.Background(), state.Context["user_id"])
		if err != nil {
			return nil, err
		}

		state.User = user
		return user, nil
	}

	// If we cannot find the user we return an error that we have no user identifier in the context
	return nil, ErrorNoUserIdentifierInContext
}

// Same as LoadUserFromContext but returns nil if no user is found
func TryLoadUserFromContext(state *model.AuthenticationSession, services *model.Repositories) (*model.User, error) {

	user, err := LoadUserFromContext(state, services)

	if err != nil && err != ErrorNoUserIdentifierInContext {
		return nil, err
	}

	return user, nil
}
