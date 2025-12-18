package model

import "net/http"

// HttpAuthContext is the context for the http authentication
type HttpAuthContext struct {

	// RequestHeaders are the headers of the incoming request
	RequestIP      string            `json:"request_ip"`
	RequestHeaders map[string]string `json:"request_headers"`
	RequestCookies map[string]string `json:"request_cookies"`

	// Response Modifications
	AdditionalResponseHeaders map[string]string      `json:"additional_response_headers"`
	AdditionalResponseCookies map[string]http.Cookie `json:"additional_response_cookies"`
}
