package oauth2

import (
	"encoding/json"
	"goiam/internal/service"

	"github.com/valyala/fasthttp"
)

// OpenIDConfiguration represents the OpenID Connect Discovery configuration
type OpenIDConfiguration struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
}

// HandleOpenIDConfiguration handles the OpenID Connect Discovery endpoint
func HandleOpenIDConfiguration(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// TODO this should be optimized to only require one service call to be more efficient
	// but currently we need the registry and flow seperatly
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("realm not found")
		return
	}

	baseURL := loadedRealm.Config.BaseUrl

	config := OpenIDConfiguration{
		Issuer:                            baseURL,
		AuthorizationEndpoint:             baseURL + "/oauth2/authorize",
		TokenEndpoint:                     baseURL + "/oauth2/token",
		UserinfoEndpoint:                  baseURL + "/oauth2/userinfo",
		JwksURI:                           baseURL + "/oauth2/jwks",
		ScopesSupported:                   []string{"openid", "profile"},
		ResponseTypesSupported:            []string{"code"},
		ResponseModesSupported:            []string{"query", "fragment"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token", "client_credentials"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"ES256"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic"},
		ClaimsSupported:                   []string{"sub", "iss", "aud", "exp", "iat", "auth_time", "nonce", "acr", "amr", "name", "given_name", "family_name", "username"},
	}

	// Marshal the configuration to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("failed to marshal configuration")
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
