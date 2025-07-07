package oauth2

import (
	"encoding/json"

	"github.com/gianlucafrei/GoAM/internal/service"

	"github.com/valyala/fasthttp"
)

// HandleTokenIntrospection handles the OAuth2 token introspection endpoint
// @Summary OAuth2 Token Introspection Endpoint
// @Description Introspects an OAuth2 token and returns information about it according to RFC 7662
// @Tags OAuth2
// @Accept x-www-form-urlencoded
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param token formData string true "The token to introspect"
// @Param token_type_hint formData string false "A hint about the type of the token submitted for introspection"
// @Success 200 {object} service.TokenIntrospectionResponse "Token introspection response"
// @Failure 400 {string} string "Bad Request - Token is required"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/oauth2/introspect [post]
func HandleTokenIntrospection(ctx *fasthttp.RequestCtx) {

	// Get token from form
	token := string(ctx.FormValue("token"))
	if token == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Token is required")
		return
	}

	// Get tenant and realm from context
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Currently the token_type_hint is not implemented
	tokenIntrospectionRequest := &service.TokenIntrospectionRequest{
		Token: token,
	}

	// Call service to introspect token
	introspectionResp, oauthErr := service.GetServices().OAuth2Service.IntrospectAccessToken(tenant, realm, tokenIntrospectionRequest)
	if oauthErr != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal server error")
		return
	}

	// Set response headers
	ctx.SetContentType("application/json")

	// Write response
	jsonData, err := json.MarshalIndent(introspectionResp, "", "  ")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal server error")
		return
	}
	ctx.SetBody(jsonData)
}
