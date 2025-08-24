package node_github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
)

var GithubLoginNode = &model.NodeDefinition{
	Name:                 "githubLogin",
	PrettyName:           "GitHub OAuth Login",
	Description:          "Handles GitHub OAuth authentication flow, including redirect to GitHub and processing the authorization code",
	Category:             "Social Login",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{""},
	PossiblePrompts:      map[string]string{"__redirect": "url", "code": "string"},
	OutputContext:        []string{"github-username", "github-access-token", "github-refresh-token", "github-token-type", "github-scope", "github-user-id", "github-avatar-url", "github-email"},
	PossibleResultStates: []string{"existing-user", "new-user", "failure"},
	CustomConfigOptions: map[string]string{
		"github-client-id":          "The client id of the Github app",
		"github-client-secret":      "The client secret of the Github app",
		"github-scope":              "The list of scopes to request from Github, comma separated",
		"create-user-if-not-exists": "If 'true' then create a new user if the user does not exist",
	},
	Run: RunGithubLoginNode,
}

func RunGithubLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	githubClientID := node.CustomConfig["github-client-id"]
	githubClientSecret := node.CustomConfig["github-client-secret"]
	githubScope := node.CustomConfig["github-scope"]

	if githubClientID == "" || githubClientSecret == "" || githubScope == "" {

		// This is a hard error as the node is misconfigured
		return model.NewNodeResultWithError(fmt.Errorf("github-client-id, github-client-secret and github-scope are required"))
	}

	code := input["code"]

	// If we have a code we exchange it for an access token, otherwise we return with a redirect to Github
	if code == "" {

		// Generate a new login with github
		redirectURL := fmt.Sprintf(
			"%s?client_id=%s&redirect_uri=%s&scope=%s",
			githubAuthURL,
			url.QueryEscape(githubClientID),
			url.QueryEscape(state.LoginUri),
			url.QueryEscape(githubScope),
		)

		return model.NewNodeResultWithPrompts(map[string]string{
			"__redirect": redirectURL,
		})
	}

	// Get the access token from Github
	githubResponse, err := getGithubAccessToken(code, githubClientID, githubClientSecret)
	if err != nil {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("failed to get github access token")
		return model.NewNodeResultWithCondition("failure")
	}

	// If the access token is not set we return with failure state
	if githubResponse.AccessToken == "" {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("failed to get github access token")
		return model.NewNodeResultWithCondition("failure")
	}

	// Get the user data from Github
	githubData, err := getGithubData(githubResponse.AccessToken)
	if err != nil {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("failed to get github user data")
		return model.NewNodeResultWithCondition("failure")
	}

	githubAttributeValue := model.GitHubAttributeValue{
		GitHubUserID:       fmt.Sprintf("%d", githubData.ID),
		GitHubRefreshToken: githubResponse.RefreshToken,
		GitHubEmail:        githubData.Email,
		GitHubAvatarURL:    githubData.AvatarURL,
		GitHubUsername:     githubData.Login,
		GitHubAccessToken:  githubResponse.AccessToken,
		GitHubTokenType:    githubResponse.TokenType,
		GitHubScope:        githubResponse.Scope,
	}

	// Store the github attribute in the context
	githubAttributeJSON, _ := json.Marshal(githubAttributeValue)
	state.Context["github"] = string(githubAttributeJSON)

	// Check if the user exists in the database
	user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeGitHub, fmt.Sprintf("%d", githubData.ID))
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// If the user does not exist we return with new user state
	if user == nil {
		return model.NewNodeResultWithCondition("new-user")
	}

	// Store the user in the state and finish
	state.User = user
	return model.NewNodeResultWithCondition("existing-user")
}
