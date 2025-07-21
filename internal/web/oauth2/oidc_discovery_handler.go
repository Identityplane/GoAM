package oauth2

import (
	"encoding/json"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/webutils"

	"github.com/valyala/fasthttp"
)

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

// HandleOpenIDConfiguration returns the OpenID Connect configuration
// @Summary Get OpenID Connect Configuration
// @Description Returns the OpenID Connect configuration for the specified realm
// @Tags OpenID Connect
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} OpenIDConfiguration
// @Failure 404 {string} string "Realm not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/oauth2/.well-known/openid-configuration [get]
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

	// If the realm does not have a base url we infer it from the request
	baseURL := loadedRealm.Config.BaseUrl
	if baseURL == "" {
		baseURL = webutils.GetFallbackUrl(ctx, tenant, realm)
	}

	config := OpenIDConfiguration{
		Issuer:                            loadedRealm.Config.BaseUrl,
		AuthorizationEndpoint:             baseURL + "/oauth2/authorize",
		TokenEndpoint:                     baseURL + "/oauth2/token",
		UserinfoEndpoint:                  baseURL + "/oauth2/userinfo",
		JwksURI:                           baseURL + "/oauth2/.well-known/jwks.json",
		ScopesSupported:                   []string{"openid", "profile"},
		ResponseTypesSupported:            []string{"code"},
		ResponseModesSupported:            []string{"query"},
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

// HandleJWKs returns the JWKs for the specified realm
// @Summary Get JWKs
// @Description Returns the JWKs for the specified realm
// @Tags OpenID Connect
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} object "JSON Web Key Set"
// @Failure 404 {string} string "Realm not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/oauth2/.well-known/jwks.json [get]
func HandleJWKs(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	jwks, err := service.GetServices().JWTService.LoadPublicKeys(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("failed to get JWKs")
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody([]byte(jwks))
}
