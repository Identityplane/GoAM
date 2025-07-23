package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// top level middleware, called before the router
func TopLevelMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		log := logger.GetLogger()

		// here we can handle request before the router
		next(ctx)

		log.Info().
			Str("method", string(ctx.Method())).
			Str("path", string(ctx.Path())).
			Int("status", ctx.Response.StatusCode()).
			Msg("request processed")
	}
}

// traceIDMiddleware adds a unique trace ID to each request
func traceIDMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		traceID := uuid.New().String()
		ctx.SetUserValue("trace_id", traceID)
		ctx.Response.Header.Set("X-Trace-ID", traceID)
		next(ctx)
	}
}

// loggingMiddleware logs incoming requests
func loggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		startTime := time.Now()
		traceID := ctx.UserValue("trace_id").(string)
		method := string(ctx.Method())
		path := string(ctx.Path())

		// Get ip address from request
		var userIP string
		if config.IsXForwardedForEnabled() {
			ip := string(ctx.Request.Header.Peek("X-Forwarded-For"))
			if ip != "" {
				userIP = ip
			}
		} else {
			userIP = ctx.RemoteIP().String()
		}
		ctx.SetUserValue("user_ip", userIP)

		// Excecute the request
		next(ctx)

		// Calculate request processing time
		duration := time.Since(startTime)
		durationMs := duration.Milliseconds()
		if config.EnableRequestTiming {
			ctx.Response.Header.Set("Server-Timing", fmt.Sprintf("req;dur=%d", durationMs))
		}

		// Log response details
		log := logger.GetLogger()
		log.Info().
			Str("trace_id", traceID).
			Int("status", ctx.Response.StatusCode()).
			Str("method", method).
			Str("path", path).
			Str("ip", userIP).
			Str("user_agent", string(ctx.UserAgent())).
			Str("referer", string(ctx.Referer())).
			Str("host", string(ctx.Host())).
			Int("duration_ms", int(durationMs)).
			Int("size_bytes", len(ctx.Response.Body())).
			Str("protocol", string(ctx.Request.URI().Scheme())).
			Msg("http response")
	}
}

// recoveryMiddleware handles panics and returns appropriate error responses
func recoveryMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				traceID := ctx.UserValue("trace_id").(string)

				// Log the panic with stack trace
				logBuf := bytes.Buffer{}
				logBuf.WriteString("[PANIC RECOVERED] ")
				logBuf.WriteString(fmt.Sprintf("TraceID: %s | Recovered: %v\n", traceID, r))
				logBuf.Write(debug.Stack())
				fmt.Println(logBuf.String())

				// Respond with JSON error
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetContentType("application/json")
				_ = json.NewEncoder(ctx).Encode(map[string]string{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()

		next(ctx)
	}
}

func setCorsHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.Response.Header.Set("Access-Control-Max-Age", "3600") // 1 hour
}

// handleOptions is a minimal handler for OPTIONS requests
func handleOptions(ctx *fasthttp.RequestCtx) {

	setCorsHeaders(ctx)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// corsMiddleware handles CORS headers
func cors(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Set CORS headers
		setCorsHeaders(ctx)
		next(ctx)
	}
}

func securityHeaders(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		next(ctx)

		ctx.Response.Header.Set("Strict-Transport-Security", "max-age=31536000;")
		ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
		ctx.Response.Header.Set("X-Frame-Options", "DENY")
		ctx.Response.Header.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// If the response has a cspNonce, we set the CSP, otherwise we set a very strict CSP as default
		cspNonce := ctx.UserValue("cspNonce")
		csp := ctx.UserValue("csp")

		if csp != nil {
			// This is used for example for the swagger UI where we need to allow the swagger UI to load the swagger.js file
			ctx.Response.Header.Set("Content-Security-Policy", csp.(string))
		} else if cspNonce != nil {
			// This is the main csp for all authentication requests
			cspNonceString := cspNonce.(string)
			csp := fmt.Sprintf("script-src 'nonce-%s' 'strict-dynamic' 'unsafe-inline' http: https:; object-src 'none'; base-uri 'none';", cspNonceString)
			ctx.Response.Header.Set("Content-Security-Policy", csp)
		} else {
			// This is the csp fallback for all api requests
			ctx.Response.Header.Set("Content-Security-Policy", "default-src 'none';")
		}

	}
}

// WrapMiddleware wraps a handler with all the necessary middleware
func WrapMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return TopLevelMiddleware(
		traceIDMiddleware(
			loggingMiddleware(
				recoveryMiddleware(
					securityHeaders(h),
				),
			),
		),
	)
}

func adminAuthMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		token := string(ctx.Request.Header.Peek("Authorization"))
		if token != "" {
			token = strings.TrimPrefix(token, "Bearer ")

			// We use the token introspection endpoint to check if the token is valid

			tenant := "internal"
			realm := "internal"

			tokenIntrospectionRequest := &service.TokenIntrospectionRequest{
				Token: token,
			}

			// Call service to introspect token
			introspectionResp, oauthErr := service.GetServices().OAuth2Service.IntrospectAccessToken(tenant, realm, tokenIntrospectionRequest)
			if oauthErr != nil {
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				ctx.SetBodyString("Invalid access token")
				return
			}

			if !introspectionResp.Active {
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				ctx.SetBodyString("Invalid access token")
				return
			}

			ctx.SetUserValue("sub", introspectionResp.Sub)
			ctx.SetUserValue("scope", introspectionResp.Scope)

			// Load the user from the database
			user, err := service.GetServices().UserService.GetUserByID(context.Background(), tenant, realm, introspectionResp.Sub)
			if err != nil {
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				ctx.SetBodyString("Invalid access token")
				return
			}

			if user == nil {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBodyString("Could not load user from token")
				return
			}

			ctx.SetUserValue("user", user)
		}

		// If the admin authz check is disabled, we skip the authz check
		if config.UnsafeDisableAdminAuthzCheck && token == "" {
			next(ctx)
			return
		}

		// Validate if the user has access to the tenant and realm
		tenantAny := ctx.UserValue("tenant")
		realmAny := ctx.UserValue("realm")

		if tenantAny != nil && realmAny != nil {

			tenant := tenantAny.(string)
			realm := realmAny.(string)

			userAny := ctx.UserValue("user")

			if userAny == nil {
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				ctx.SetBodyString("Unauthorized")
				return
			}

			user := userAny.(*model.User)

			hasAccess, _ := service.GetServices().AdminAuthzService.CheckAccess(user, tenant, realm, "")
			if !hasAccess {
				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetBodyString("No access to tenant or realm")
				return
			}
		}

		next(ctx)
	}
}

func adminMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {

	return WrapMiddleware(
		cors(
			adminAuthMiddleware(
				next,
			),
		),
	)
}
