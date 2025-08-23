package node_utils

import "github.com/Identityplane/GoAM/pkg/model"

// GetAccountNameFromContext returns the display name of the user from the context
// This name is ised in various context where we need to display the id to the user
func GetAccountNameFromContext(state *model.AuthenticationSession) string {

	// If there us no user in the context we return an empty string
	if state.User == nil {
		return ""
	}

	// If we have a username we use that
	if state.User.Username != "" {
		return state.User.Username
	}

	// If we have a email we use that
	if state.User.Email != "" {
		return state.User.Email
	}

	// If we have a phone we use that
	if state.User.Phone != "" {
		return state.User.Phone
	}

	// Otherwise we return the id
	return state.User.ID
}
