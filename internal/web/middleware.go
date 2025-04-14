package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func recoverMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		traceID := uuid.New().String()
		ctx.SetUserValue("trace_id", traceID)

		// Add it to the response header for every request
		ctx.Response.Header.Set("X-Trace-ID", traceID)

		// Log the request
		method := string(ctx.Method())
		path := string(ctx.Path())
		fmt.Printf("[REQ] %s %s | trace_id=%s\n", method, path, traceID)

		defer func() {
			if r := recover(); r != nil {
				// Log it with stack trace
				logBuf := bytes.Buffer{}
				logBuf.WriteString("[PANIC RECOVERED] ")
				logBuf.WriteString(fmt.Sprintf("TraceID: %s | Recovered: %v\n", traceID, r))
				logBuf.Write(debug.Stack())
				fmt.Println(logBuf.String())

				// Respond with JSON error + trace ID
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

func WrapMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return recoverMiddleware(h)
}

// Middleware-style JSON handler
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
