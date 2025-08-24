package node_github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Identityplane/GoAM/pkg/model"
)

var GithubCreateUserNode = &model.NodeDefinition{
	Name:                 "githubCreateUser",
	PrettyName:           "Create GitHub User",
	Description:          "Creates a new user account using information from GitHub OAuth authentication",
	Category:             "User Management",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"github"},
	PossibleResultStates: []string{"created"},
	OutputContext:        []string{"user"},
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

func RunGithubCreateUserNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// We we have no github attribute in the context we return an error
	if state.Context["github"] == "" {
		return model.NewNodeResultWithError(fmt.Errorf("github is required in state context"))
	}

	// Parse the github attribute
	githubAttribute := model.GitHubAttributeValue{}
	err := json.Unmarshal([]byte(state.Context["github"]), &githubAttribute)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to parse github attribute: %w", err))
	}

	// Creata a new user in the context if it does not exist
	if state.User == nil {
		state.User = &model.User{
			Status: "active",
		}
	}

	// Append the github attribute to the user with the user id as the index
	state.User.AddAttribute(&model.UserAttribute{
		Index: githubAttribute.GitHubUserID,
		Type:  model.AttributeTypeGitHub,
		Value: githubAttribute,
	})

	// Save the user
	err = services.UserRepo.CreateOrUpdate(context.Background(), state.User)
	if err != nil {
		return model.NewNodeResultWithError(fmt.Errorf("failed to save user: %w", err))
	}

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
