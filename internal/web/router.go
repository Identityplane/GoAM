package web

import (
	"encoding/json"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/web/admin_api"
	"github.com/Identityplane/GoAM/internal/web/auth"
	"github.com/Identityplane/GoAM/internal/web/debug"
	"github.com/Identityplane/GoAM/internal/web/oauth2"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var redirectUrl string

func New() *router.Router {
	r := router.New()

	// Admin routes
	admin := r.Group("/admin")
	admin.OPTIONS("/{name:*}", WrapMiddleware(handleOptions)) // Cors for options requests requests

	admin.GET("/whoami", adminMiddlewareAllowsAll(admin_api.HandleWhoAmI))
	admin.GET("/realms", adminMiddlewareAllowsAll(admin_api.HandleListRealms))
	admin.GET("/system/stats", adminMiddlewareAllowsAll(admin_api.HandleSystemStats))
	admin.GET("/tenants/check-availability/{tenant_name}", adminMiddlewareAllowsAll(admin_api.HandleTenantNameAvailable))
	admin.POST("/tenants", adminMiddlewareAllowsAll(admin_api.HandleCreateTenant))

	admin.GET("/{tenant}/{realm}/users", adminMiddleware(admin_api.HandleListUsers))
	admin.GET("/{tenant}/{realm}/users/stats", adminMiddleware(admin_api.HandleGetUserStats))
	admin.GET("/{tenant}/{realm}/users/{id}", adminMiddleware(admin_api.HandleGetUser))
	admin.POST("/{tenant}/{realm}/users/{id}", adminMiddleware(admin_api.HandleCreateUser))
	admin.PUT("/{tenant}/{realm}/users/{id}", adminMiddleware(admin_api.HandleUpdateUser))
	admin.DELETE("/{tenant}/{realm}/users/{id}", adminMiddleware(admin_api.HandleDeleteUser))

	// User attribute management routes
	admin.GET("/{tenant}/{realm}/users/{id}/attributes", adminMiddleware(admin_api.HandleListUserAttributes))
	admin.POST("/{tenant}/{realm}/users/{id}/attributes", adminMiddleware(admin_api.HandleCreateUserAttribute))
	admin.GET("/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id}", adminMiddleware(admin_api.HandleGetUserAttribute))
	admin.PATCH("/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id}", adminMiddleware(admin_api.HandleUpdateUserAttribute))
	admin.DELETE("/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id}", adminMiddleware(admin_api.HandleDeleteUserAttribute))

	admin.GET("/{tenant}/{realm}/dashboard", adminMiddleware(admin_api.HandleDashboard))

	admin.GET("/{tenant}/{realm}/", adminMiddleware(admin_api.HandleGetRealm))
	admin.POST("/{tenant}/{realm}/", adminMiddleware(admin_api.HandleCreateRealm))
	admin.PATCH("/{tenant}/{realm}/", adminMiddleware(admin_api.HandleUpdateRealm))
	admin.DELETE("/{tenant}/{realm}/", adminMiddleware(admin_api.HandleDeleteRealm))

	// Application management routes
	admin.GET("/{tenant}/{realm}/applications", adminMiddleware(admin_api.HandleListApplications))
	admin.GET("/{tenant}/{realm}/applications/{client_id}", adminMiddleware(admin_api.HandleGetApplication))
	admin.POST("/{tenant}/{realm}/applications/{client_id}", adminMiddleware(admin_api.HandleCreateApplication))
	admin.PUT("/{tenant}/{realm}/applications/{client_id}", adminMiddleware(admin_api.HandleUpdateApplication))
	admin.DELETE("/{tenant}/{realm}/applications/{client_id}", adminMiddleware(admin_api.HandleDeleteApplication))
	admin.POST("/{tenant}/{realm}/applications/{client_id}/regenerate-secret", adminMiddleware(admin_api.HandleRegenerateClientSecret))

	// Flow management routes
	admin.GET("/{tenant}/{realm}/flows", adminMiddleware(admin_api.HandleListFlows))
	admin.GET("/{tenant}/{realm}/flows/{flow}", adminMiddleware(admin_api.HandleGetFlow))
	admin.POST("/{tenant}/{realm}/flows/{flow}", adminMiddleware(admin_api.HandleCreateFlow))
	admin.PATCH("/{tenant}/{realm}/flows/{flow}", adminMiddleware(admin_api.HandleUpdateFlow))
	admin.DELETE("/{tenant}/{realm}/flows/{flow}", adminMiddleware(admin_api.HandleDeleteFlow))

	// Flow defintion routes
	admin.POST("/{tenant}/{realm}/flows/validate", adminMiddleware(admin_api.HandleValidateFlowDefinition))
	admin.GET("/{tenant}/{realm}/flows/{flow}/definition", adminMiddleware(admin_api.HandleGetFlowDefintion))
	admin.PUT("/{tenant}/{realm}/flows/{flow}/definition", adminMiddleware(admin_api.HandlePutFlowDefintion))

	// Node management routes
	admin.GET("/nodes", adminMiddleware(admin_api.HandleListNodes))

	// Debug routes
	r.GET("/{tenant}/{realm}/debug/{flow}/graph.svg", adminMiddleware(debug.HandleFlowGraphSVG))

	// Static files
	r.GET("/{tenant}/{realm}/static/{filename}", WrapMiddleware(StaticHandler))
	r.GET("/{tenant}/{realm}/assets/{filename}", WrapMiddleware(auth.HandleStaticAssets))

	// Health endpoints
	r.GET("/healthz", WrapMiddleware(handleLiveness))
	r.GET("/readyz", WrapMiddleware(handleReadiness))
	r.GET("/info", WrapMiddleware(handleInfo))

	// Swagger UI
	r.GET("/swagger/", WrapMiddleware(HandleSwaggerUI))
	r.GET("/swagger/{*path}", WrapMiddleware(HandleSwaggerUI))

	// Main authentication routes
	r.GET("/{tenant}/{realm}/auth/{path}", WrapMiddleware(auth.HandleAuthRequest))
	r.POST("/{tenant}/{realm}/auth/{path}", WrapMiddleware(auth.HandleAuthRequest))

	// Oauth + OIDC
	r.GET("/{tenant}/{realm}/oauth2/authorize", WrapMiddleware(oauth2.HandleAuthorizeEndpoint))
	r.GET("/{tenant}/{realm}/oauth2/finishauthorize", WrapMiddleware(oauth2.FinsishOauth2AuthorizationEndpoint))

	r.GET("/{tenant}/{realm}/oauth2/.well-known/openid-configuration", cors(WrapMiddleware(oauth2.HandleOpenIDConfiguration)))
	r.POST("/{tenant}/{realm}/oauth2/token", cors(WrapMiddleware(oauth2.HandleTokenEndpoint)))

	// OIDC Userinfo endpoint
	r.GET("/{tenant}/{realm}/oauth2/userinfo", cors(WrapMiddleware(oauth2.HandleUserinfoEndpoint)))
	r.POST("/{tenant}/{realm}/oauth2/userinfo", cors(WrapMiddleware(oauth2.HandleUserinfoEndpoint)))
	r.OPTIONS("/{tenant}/{realm}/oauth2/userinfo", WrapMiddleware(handleOptions))

	// OAuth 2 Token Introspection endpoint
	r.POST("/{tenant}/{realm}/oauth2/introspect", cors(WrapMiddleware(oauth2.HandleTokenIntrospection)))

	// OIDC JWKS endpoint
	r.GET("/{tenant}/{realm}/oauth2/.well-known/jwks.json", cors(WrapMiddleware(oauth2.HandleJWKs)))

	// handleNotFound is the fallback handler for unmatched routes
	redirectUrl = config.ServerSettings.NotFoundRedirectUrl
	r.NotFound = WrapMiddleware(func(ctx *fasthttp.RequestCtx) {

		if redirectUrl != "" {
			ctx.Redirect(redirectUrl, fasthttp.StatusSeeOther)
			return
		}

		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "not found",
			"path":  string(ctx.Path()),
		})
	})

	return r
}
