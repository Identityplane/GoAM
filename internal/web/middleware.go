package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// top level middleware, called before the router
func TopLevelMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		// We need to handle OPTIONS requests here, because the router doesn't handle them
		if string(ctx.Method()) == "OPTIONS" {
			handleOptions(ctx)
			return
		}

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

		// Also log the response status code
		fmt.Printf("[RES] %d %s %s | trace_id=%s", ctx.Response.StatusCode(), method, path, traceID)
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
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.Response.Header.Set("Access-Control-Max-Age", "86400") // 24 hours
}

// handleOptions is a minimal handler for OPTIONS requests
func handleOptions(ctx *fasthttp.RequestCtx) {
	setCorsHeaders(ctx)
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// corsMiddleware handles CORS headers
func corsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Set CORS headers
		setCorsHeaders(ctx)
		next(ctx)
	}
}

// WrapMiddleware wraps a handler with all necessary middleware
func WrapMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return traceIDMiddleware(
		loggingMiddleware(
			recoveryMiddleware(
				corsMiddleware(h),
			),
		),
	)
}

// jsonHandler is a middleware-style JSON handler
func jsonHandler(f func(*fasthttp.RequestCtx) any) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		result := f(ctx)

		switch v := result.(type) {
		case error:
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			_ = json.NewEncoder(ctx).Encode(map[string]string{"error": v.Error()})
		case map[string]any, map[string]string:
			ctx.SetStatusCode(fasthttp.StatusOK)
			_ = json.NewEncoder(ctx).Encode(v)
		default:
			ctx.SetStatusCode(fasthttp.StatusOK)
			_ = json.NewEncoder(ctx).Encode(map[string]any{"data": v})
		}
	}
}
