package oauth2

import (
	"encoding/base64"
	"testing"

	"github.com/Identityplane/GoAM/internal/lib/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestGetClientAuthenticationFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(*fasthttp.RequestCtx)
		expectedResult *oauth2.Oauth2ClientAuthentication
		description    string
	}{
		{
			name: "Basic Auth Header - Valid",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Create valid Basic Auth header: "client_id:client_secret" base64 encoded
				credentials := "test_client:test_secret"
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "test_client",
				ClientSecret: "test_secret",
			},
			description: "Should extract client credentials from valid Basic Auth header",
		},
		{
			name: "Basic Auth Header - Invalid Base64",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				ctx.Request.Header.Set("Authorization", "Basic invalid-base64")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "",
				ClientSecret: "",
			},
			description: "Should return empty strings when Basic Auth header contains invalid base64",
		},
		{
			name: "Basic Auth Header - Invalid Format",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Create Basic Auth header with wrong format (no colon separator)
				credentials := "invalid_format"
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "",
				ClientSecret: "",
			},
			description: "Should return empty strings when Basic Auth header has invalid format (no colon separator)",
		},
		{
			name: "Basic Auth Header - Empty Credentials",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				credentials := ":"
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "",
				ClientSecret: "",
			},
			description: "Should handle empty client ID and secret from Basic Auth header",
		},
		{
			name: "Form Body - Valid",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Set form body with client credentials
				formData := "client_id=form_client&client_secret=form_secret"
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "form_client",
				ClientSecret: "form_secret",
			},
			description: "Should extract client credentials from form body when no Basic Auth header",
		},
		{
			name: "Form Body - Missing Fields",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Set form body without client credentials
				formData := "grant_type=client_credentials&scope=read"
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "",
				ClientSecret: "",
			},
			description: "Should return empty strings when form body doesn't contain client credentials",
		},
		{
			name: "Form Body - Invalid URL Encoding",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Set invalid form body
				ctx.Request.SetBodyString("invalid%form%data")
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "",
				ClientSecret: "",
			},
			description: "Should return empty strings when form body has invalid URL encoding",
		},
		{
			name: "No Authentication",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Don't set any authentication headers or body
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "",
				ClientSecret: "",
			},
			description: "Should return empty strings when no authentication is provided",
		},
		{
			name: "Basic Auth Header - Special Characters",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Test with special characters in credentials
				credentials := "client@domain.com:secret@123"
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "client@domain.com",
				ClientSecret: "secret@123",
			},
			description: "Should handle special characters in Basic Auth credentials",
		},
		{
			name: "Form Body - Special Characters",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Test with special characters in form data
				formData := "client_id=client%40domain.com&client_secret=secret%40123"
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "client@domain.com",
				ClientSecret: "secret@123",
			},
			description: "Should handle URL-encoded special characters in form body",
		},
		{
			name: "Basic Auth Header - Multiple Colons",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Test with multiple colons in credentials
				credentials := "client:secret:extra"
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "client",
				ClientSecret: "secret:extra",
			},
			description: "Should handle multiple colons in Basic Auth credentials (first colon is separator)",
		},
		{
			name: "Authorization Header Without Basic Prefix",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Set Authorization header without "Basic " prefix
				ctx.Request.Header.Set("Authorization", "Bearer some-token")

				// Should fall back to form body parsing
				formData := "client_id=fallback_client&client_secret=fallback_secret"
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "fallback_client",
				ClientSecret: "fallback_secret",
			},
			description: "Should fall back to form body when Authorization header doesn't have Basic prefix",
		},
		{
			name: "Empty Authorization Header",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Set empty Authorization header
				ctx.Request.Header.Set("Authorization", "")

				// Should fall back to form body parsing
				formData := "client_id=form_client&client_secret=form_secret"
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "form_client",
				ClientSecret: "form_secret",
			},
			description: "Should fall back to form body when Authorization header is empty",
		},
		{
			name: "Basic Auth Header with Whitespace",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Test with whitespace in credentials
				credentials := " client_id : client_secret "
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     " client_id ",
				ClientSecret: " client_secret ",
			},
			description: "Should handle whitespace in Basic Auth credentials",
		},
		{
			name: "Form Body with Whitespace",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Test with whitespace in form data
				formData := "client_id= spaced_client &client_secret= spaced_secret "
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     " spaced_client ",
				ClientSecret: " spaced_secret ",
			},
			description: "Should handle whitespace in form body credentials",
		},
		{
			name: "Basic Auth Takes Priority Over Form Body",
			setupRequest: func(ctx *fasthttp.RequestCtx) {
				// Set both Basic Auth header and form body
				credentials := "basic_client:basic_secret"
				encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
				ctx.Request.Header.Set("Authorization", "Basic "+encoded)

				formData := "client_id=form_client&client_secret=form_secret"
				ctx.Request.SetBodyString(formData)
				ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			},
			expectedResult: &oauth2.Oauth2ClientAuthentication{
				ClientID:     "basic_client",
				ClientSecret: "basic_secret",
			},
			description: "Basic Auth should take priority over form body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request context
			ctx := &fasthttp.RequestCtx{}
			ctx.Init(&ctx.Request, nil, nil)

			// Setup the request according to the test case
			tt.setupRequest(ctx)

			// Call the function under test
			result := getClientAuthenticationFromRequest(ctx)

			// Assert the result
			assert.NotNil(t, result, "Function should always return a result")
			assert.Equal(t, tt.expectedResult.ClientID, result.ClientID, "ClientID mismatch: "+tt.description)
			assert.Equal(t, tt.expectedResult.ClientSecret, result.ClientSecret, "ClientSecret mismatch: "+tt.description)
		})
	}
}
