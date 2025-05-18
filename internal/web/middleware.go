package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"goiam/internal/logger"
	"goiam/internal/service"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// top level middleware, called before the router
func TopLevelMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		// here we can handle request before the router
		next(ctx)
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
		traceID := ctx.UserValue("trace_id").(string)
		method := string(ctx.Method())
		path := string(ctx.Path())

		next(ctx)

		// Log response details
		logger.InfoWithFields(traceID, "HTTP Response", map[string]interface{}{
			"status":      ctx.Response.StatusCode(),
			"method":      method,
			"path":        path,
			"ip":          ctx.RemoteIP().String(),
			"user_agent":  string(ctx.UserAgent()),
			"referer":     string(ctx.Referer()),
			"host":        string(ctx.Host()),
			"duration_ms": ctx.Time().Sub(ctx.ConnTime()).Milliseconds(),
			"size_bytes":  len(ctx.Response.Body()),
			"protocol":    string(ctx.Request.URI().Scheme()),
		})
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

			ctx.SetUserValue("user", user)
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
