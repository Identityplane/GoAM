package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/model"

	"github.com/google/uuid"
)

var GithubLoginNode = &NodeDefinition{
	Name:                 "githubLogin",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{""},
	PossiblePrompts:      map[string]string{"__redirect": "url", "code": "string"},
	OutputContext:        []string{"github-username", "github-access-token", "github-refresh-token", "github-token-type", "github-scope", "github-user-id", "github-avatar-url", "github-email"},
	PossibleResultStates: []string{"existing-user", "new-user", "failure"},
	CustomConfigOptions:  []string{"github-client-id", "github-client-secret", "github-scope"},
	Run:                  RunGithubLoginNode,
}

var GithubCreateUserNode = &NodeDefinition{
	Name:                 "githubCreateUser",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"github-username"},
	PossibleResultStates: []string{"created"},
	OutputContext:        []string{"github-username", "user"},
	Run:                  RunGithubCreateUserNode,
}

const (
	githubProvider = "github"
	githubAuthURL  = "https://github.com/login/oauth/authorize"
)

var (
	githubTokenURL = "https://github.com/login/oauth/access_token"
	githubUserURL  = "https://api.github.com/user"
)

func RunGithubLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

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

	githubDataJSON, _ := json.Marshal(githubData)

	state.Context["github-access-token"] = githubResponse.AccessToken
	state.Context["github-refresh-token"] = githubResponse.RefreshToken
	state.Context["github-token-type"] = githubResponse.TokenType
	state.Context["github-scope"] = githubResponse.Scope
	state.Context["github-user-id"] = fmt.Sprintf("%d", githubData.ID)
	state.Context["github-username"] = githubData.Login
	state.Context["github-avatar-url"] = githubData.AvatarURL
	state.Context["github-email"] = githubData.Email
	state.Context["github-json"] = string(githubDataJSON)

	// Check if the user exists in the database
	user, err := services.UserRepo.GetByFederatedIdentifier(context.Background(), githubProvider, fmt.Sprintf("%d", githubData.ID))
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

func RunGithubCreateUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {

	// if we have a email in the context we use that, otherwise we use the github email
	email := state.Context["email"]
	if email == "" {
		email = state.Context["github-email"]
	}

	// if we have a username in the context we use that, otherwise we use the github username
	username := state.Context["username"]
	if username == "" {
		username = state.Context["github-username"]
	}

	githubUserID := state.Context["github-user-id"]

	if githubUserID == "" {
		return model.NewNodeResultWithError(fmt.Errorf("github-user-id is required in state context"))
	}

	// if we have a avatar url in the context we use that, otherwise we use the github avatar url
	avatarURL := state.Context["avatar-url"]
	if avatarURL == "" {
		avatarURL = state.Context["github-avatar-url"]
	}

	if username == "" {
		return model.NewNodeResultWithError(fmt.Errorf("username or github-username is required in state context"))
	}

	githubProviderString := "github"

	// Create a new user
	user := &model.User{
		ID:                uuid.NewString(),
		Username:          username,
		Email:             email,
		ProfilePictureURI: avatarURL,
		FederatedIDP:      &githubProviderString,
		FederatedID:       &githubUserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Create the user
	if err := services.UserRepo.Create(context.Background(), user); err != nil {
		return model.NewNodeResultWithError(err)
	}

	state.User = user
	return model.NewNodeResultWithCondition("created")
}

type GitHubUser struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	UserViewType      string `json:"user_view_type"`
	SiteAdmin         bool   `json:"site_admin"`
	Name              string `json:"name"`
	Company           string `json:"company"`
	Blog              string `json:"blog"`
	Location          string `json:"location"`
	Email             string `json:"email"`
	Hireable          bool   `json:"hireable"`
	Bio               string `json:"bio"`
	TwitterUsername   string `json:"twitter_username"`
	NotificationEmail string `json:"notification_email"`
	PublicRepos       int    `json:"public_repos"`
	PublicGists       int    `json:"public_gists"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// Represents the response received from Github
type githubAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func getGithubAccessToken(code, clientID, clientSecret string) (*githubAccessTokenResponse, error) {
	// Set us the request body as JSON
	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	requestJSON, _ := json.Marshal(requestBodyMap)

	// POST request to set URL
	req, reqerr := http.NewRequest(
		"POST",
		githubTokenURL,
		bytes.NewBuffer(requestJSON),
	)
	if reqerr != nil {
		return nil, reqerr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Get the response
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return nil, resperr
	}

	// Response body converted to stringified JSON
	respbody, _ := io.ReadAll(resp.Body)

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	// Return the access token (as the rest of the
	// details are relatively unnecessary for us)
	return &ghresp, nil
}

func getGithubData(accessToken string) (*GitHubUser, error) {
	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		githubUserURL,
		nil,
	)
	if reqerr != nil {
		return nil, reqerr
	}

	// Set the Authorization header before sending the request
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return nil, resperr
	}

	// Read the response as a byte slice
	respbody, _ := io.ReadAll(resp.Body)

	// Convert byte slice to string and return
	var githubUser GitHubUser
	json.Unmarshal(respbody, &githubUser)

	return &githubUser, nil
}
