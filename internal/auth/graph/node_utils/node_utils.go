package node_utils

import "github.com/Identityplane/GoAM/pkg/model"

// GetAccountNameFromContext returns the display name of the user from the context
// This name is ised in various context where we need to display the id to the user
func GetAccountNameFromContext(state *model.AuthenticationSession) string {

	// If there us no user in the context we return an empty string
	if state.User == nil {
		return ""
	}

	// TODO we need to get the display name from the user attributes

	// Otherwise we return the id
	return state.User.ID
}
