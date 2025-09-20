package node_github

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONFIG_GITHUB_CLIENT_ID          = "github-client-id"
	CONFIG_GITHUB_CLIENT_SECRET      = "github-client-secret"
	CONFIG_GITHUB_SCOPE              = "github-scope"
	CONFIG_CREATE_USER_IF_NOT_EXISTS = "github-create-user-if-not-exists"
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
		CONFIG_GITHUB_CLIENT_ID:          "The client id of the Github app",
		CONFIG_GITHUB_CLIENT_SECRET:      "The client secret of the Github app",
		CONFIG_GITHUB_SCOPE:              "The list of scopes to request from Github, comma separated",
		CONFIG_CREATE_USER_IF_NOT_EXISTS: "If 'true' then create a new user if the user does not exist",
	},
	Run: RunGithubLoginNode,
}

func RunGithubLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	githubClientID := node.CustomConfig[CONFIG_GITHUB_CLIENT_ID]
	githubClientSecret := node.CustomConfig[CONFIG_GITHUB_CLIENT_SECRET]
	githubScope := node.CustomConfig[CONFIG_GITHUB_SCOPE]

	if githubClientID == "" || githubClientSecret == "" || githubScope == "" {

		// This is a hard error as the node is misconfigured
		return model.NewNodeResultWithError(fmt.Errorf(CONFIG_GITHUB_CLIENT_ID + ", " + CONFIG_GITHUB_CLIENT_SECRET + " and " + CONFIG_GITHUB_SCOPE + " are required"))
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

	// Check if the user exists in the database by checking the github user id
	user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeGitHub, fmt.Sprintf("%d", githubData.ID))
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	if user != nil {

		// Store the user in the state and finish
		// (TODO user is not updated wiht new github info)
		state.User = user
		return model.NewNodeResultWithCondition("existing-user")
	}

	// If we reach here we have a new user that is not linked yet
	if state.User == nil {
		state.User, err = services.UserRepo.NewUserModel(state)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
	}

	// Add the github attribute to the user
	state.User.AddAttribute(&model.UserAttribute{
		Index: lib.StringPtr(fmt.Sprintf("%d", githubData.ID)),
		Type:  model.AttributeTypeGitHub,
		Value: githubAttributeValue,
	})

	// If the create user option is enabled we create the user
	if node.CustomConfig[CONFIG_CREATE_USER_IF_NOT_EXISTS] == "true" {
		err := services.UserRepo.Create(context.Background(), state.User)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
	}

	return model.NewNodeResultWithCondition("new-user")
}
