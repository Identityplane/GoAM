package node_oidc

import (
	"context"
	"fmt"
	"strings"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/pkg/model/attributes"
	oidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	CONFIG_ISSUER        = "issuer"
	CONFIG_CLIENT_ID     = "client_id"
	CONFIG_CLIENT_SECRET = "client_secret"
	CONFIG_REDIRECT_URI  = "redirect_uri"
	CONFIG_SCOPE         = "scope"
	CONFIG_CLAIMS        = "claims"

	CONDITION_OIDC_FAILURE       = "oidc_failure"
	CONDITION_OIDC_NEW_USER      = "oidc_new_user"
	CONDITION_OIDC_EXISTING_USER = "oidc_existing_user"
)

// GenericOIDCLoginNode is a node that logs in a user using a generic OIDC provider like
// Google, Apple, Microsoft, etc.
// It handles the entire OIDC flow and returns back to GoAM to sign in or create a user
// If successfull the node creates a oidc attribute for the user and also looks up if the user already exists
//
// OIDC Details:
// This node uses the userinfo endpoint to get the user's claims instead of the id_token.
// this is because this flow is more secure and the user info endpoint can contain additional claims that are not included in the id_token
// This node also uses the oidc server metadata to discover the endpoints and the supported algorithms
// OIDC serviers without the .well-known/openid-configuration endpoint are not supported at this time but will
// be added in the future with a custom config via the node config.
var GenericOIDCLoginNode = &model.NodeDefinition{
	Name:                 "genericOIDCLogin",
	PrettyName:           "Generic OIDC Login",
	Description:          "Logs in a user using a generic OIDC provider.",
	Category:             "OIDC",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{"oidcLoginResult"},
	PossibleResultStates: []string{CONDITION_OIDC_FAILURE, CONDITION_OIDC_NEW_USER, CONDITION_OIDC_EXISTING_USER},
	PossiblePrompts: map[string]string{
		"__redirect": "The redirect url to the OIDC provider",
		"code":       "The code from the OIDC provider",
		"state":      "The state from the OIDC provider",
	},
	Run: RunGenericOIDCLoginNode,
}

func RunGenericOIDCLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	ctx := context.Background()

	oauth2Config, provider, issuer, err := getOauth2Config(ctx, node, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 config: %w", err)
	}

	// if we have a code in the input we exchange it for an access token
	code := input["code"]
	if code != "" {
		oidcAttributeValue, err := finishOidcLogin(ctx, state, node, oauth2Config, provider, input, issuer)
		if err != nil {
			state.Context["oidc_error"] = err.Error()
			return model.NewNodeResultWithCondition(CONDITION_OIDC_FAILURE)
		}
		return initUserAndFinish(ctx, state, oidcAttributeValue, services)
	}

	// otherwise we perform the login
	return performOidcLogin(ctx, state, node, oauth2Config, provider)
}

func performOidcLogin(ctx context.Context, state *model.AuthenticationSession, node *model.GraphNode, oauth2Config *oauth2.Config, provider *oidc.Provider) (*model.NodeResult, error) {

	// Generate a random state and safe it to the state
	randomState := lib.GenerateSecureSessionID()
	state.Context["oidc_state"] = randomState

	// Ensure we remember the same redirect url as in the authorize request
	state.Context["oidc_redirect_url"] = oauth2Config.RedirectURL

	// Get the redirect url
	authCodeUrl := oauth2Config.AuthCodeURL(randomState)

	// return the redirect
	return model.NewNodeResultWithPrompts(map[string]string{
		"__redirect": authCodeUrl,
	})
}

func finishOidcLogin(ctx context.Context, state *model.AuthenticationSession, node *model.GraphNode, oauth2Config *oauth2.Config, provider *oidc.Provider, input map[string]string, issuer string) (*attributes.OidcAttributeValue, error) {

	code := input["code"]
	stateValue := input["state"]

	// Check if the state is valid
	if stateValue != state.Context["oidc_state"] {
		return nil, fmt.Errorf("invalid state")
	}

	// Load the redirect url from the state
	oauth2Config.RedirectURL = state.Context["oidc_redirect_url"]
	if oauth2Config.RedirectURL == "" {
		return nil, fmt.Errorf("redirect url is required")
	}

	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("state", stateValue),
	}

	// Exchange the code for an access token
	oauth2Token, err := oauth2Config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	oidcAttributeValue, err := getOidcAttributeValue(oauth2Token, oauth2Config, provider, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to get oidc attribute value: %w", err)
	}

	return oidcAttributeValue, nil
}

func initUserAndFinish(ctx context.Context, state *model.AuthenticationSession, oidcAttributeValue *attributes.OidcAttributeValue, services *model.Repositories) (*model.NodeResult, error) {

	// Check if we already have a user with the same oidc attribute value
	index := oidcAttributeValue.GetIndex()
	user, err := services.UserRepo.GetByAttributeIndex(ctx, model.AttributeTypeOidc, index)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by index: %w", err)
	}

	// If the user is not known we init a new user in the context and add the oidc attribute
	if user == nil {
		user, err = services.UserRepo.NewUserModel(state)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		user.AddAttribute(&model.UserAttribute{
			Index: &index,
			Type:  model.AttributeTypeOidc,
			Value: oidcAttributeValue,
		})

		state.User = user
		return model.NewNodeResultWithCondition(CONDITION_OIDC_NEW_USER)

	} else {
		// If the user already exists we dont update it but we return the user
		// this is to avoid any unintended side effects if the user attributes are changed on the federated provider
		// TODO: add a setting to enable this behavior so that user profile and claims are updated on the federated provider
		// as well as the acccess token and refresh token
		state.User = user
		return model.NewNodeResultWithCondition(CONDITION_OIDC_EXISTING_USER)
	}

}

func getOauth2Config(ctx context.Context, node *model.GraphNode, state *model.AuthenticationSession) (*oauth2.Config, *oidc.Provider, string, error) {

	// Get the provider
	provider, issuer, err := getProvider(ctx, node)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Get client id from config
	clientId := node.CustomConfig[CONFIG_CLIENT_ID]
	if clientId == "" {
		return nil, nil, "", fmt.Errorf("client id is required")
	}

	// Get client secret from config
	clientSecret := node.CustomConfig[CONFIG_CLIENT_SECRET]
	if clientSecret == "" {
		return nil, nil, "", fmt.Errorf("client secret is required")
	}

	redirectUrl := getRedirectUrl(node, state)
	if redirectUrl == "" {
		return nil, nil, "", fmt.Errorf("redirect url is required")
	}

	// Configure an OpenID Connect aware OAuth2 client.
	oauth2Config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &oauth2Config, provider, issuer, nil
}

func getProvider(ctx context.Context, node *model.GraphNode) (*oidc.Provider, string, error) {

	// Get issuer from config
	issuer := node.CustomConfig[CONFIG_ISSUER]
	if issuer == "" {
		return nil, "", fmt.Errorf("issuer is required")
	}

	// Create provider
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, issuer, nil
}

func getRedirectUrl(node *model.GraphNode, state *model.AuthenticationSession) string {

	// If we have a redirect url in the custom config we use that
	redirectUrl := node.CustomConfig[CONFIG_REDIRECT_URI]
	if redirectUrl != "" {
		return redirectUrl
	}

	// Otherwise we use the base url from the state
	return state.LoginUriNext
}

// Get the oidc attribute value from the oauth2 token and config
func getOidcAttributeValue(oauth2Token *oauth2.Token, oauth2Config *oauth2.Config, provider *oidc.Provider, issuer string) (*attributes.OidcAttributeValue, error) {

	userInfo, err := provider.UserInfo(context.Background(), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// The scope is wither the extra scope form the oauth2 token response or the oauth2 config scopes
	scopes := []string{}
	scope := oauth2Token.Extra("scope")
	if scope != nil {
		scopeString, ok := scope.(string)
		if !ok {
			return nil, fmt.Errorf("scope is not a string")
		}
		scopes = strings.Split(scopeString, " ")
	} else {
		scopes = oauth2Config.Scopes
	}

	claims := make(map[string]interface{})
	err = userInfo.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info claims: %w", err)
	}

	return &attributes.OidcAttributeValue{
		Issuer:       issuer,
		ClientId:     oauth2Config.ClientID,
		Scope:        strings.Join(scopes, " "),
		AccessToken:  oauth2Token.AccessToken,
		RefreshToken: oauth2Token.RefreshToken,
		ExpiresIn:    oauth2Token.ExpiresIn,
		Sub:          userInfo.Subject,
		Claims:       claims,
	}, nil
}
