package web

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

// ErrorResponse sends a JSON error response with the given message and status code
func ErrorResponse(ctx *fasthttp.RequestCtx, message string, statusCode int) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")

	errorResponse := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	jsonData, _ := json.Marshal(errorResponse)
	ctx.SetBody(jsonData)
}
