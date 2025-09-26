package model

import "net/http"

const (
	GRANT_SIMPLE_AUTH_BODY   = "simple-body"
	GRANT_SIMPLE_AUTH_COOKIE = "simple-cookie"
)

type SimpleAuthRequest struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	Grant        string `json:"grant"`
	ResponseType string `json:"response_type"`
}

type SimpleAuthContext struct {
	Request     *SimpleAuthRequest `json:"request"`
	RedirectURI string             `json:"redirect_uri"`
}

type SimpleAuthResponse struct {
	Success               bool                   `json:"success"`
	AccessToken           string                 `json:"access_token,omitempty"`
	TokenType             string                 `json:"token_type,omitempty"`
	RefreshToken          string                 `json:"refresh_token,omitempty"`
	UserClaims            map[string]interface{} `json:"user,omitempty"`
	ExpiresIn             int                    `json:"expires_in,omitempty"`
	RefreshTokenExpiresIn int                    `json:"refresh_token_expires_in,omitempty"`
	Scope                 string                 `json:"scope,omitempty"`
	Error                 string                 `json:"error,omitempty"`
}

type AuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	HttpStatusCode   int    `json:"http_status_code,omitempty"`
}

func SimpleAuthErrorFound() *AuthError {
	return &AuthError{
		Error:            "NOT_FOUND",
		ErrorDescription: "Not found",
		HttpStatusCode:   http.StatusNotFound,
	}
}

func SimpleAuthErrorInvalidClientID() *AuthError {
	return &AuthError{
		Error:            "INVALID_REQUEST",
		ErrorDescription: "Invalid client id",
		HttpStatusCode:   http.StatusBadRequest,
	}
}

func SimpleAuthErrorClientUnauthorized() *AuthError {
	return &AuthError{
		Error:            "INVALID_REQUEST",
		ErrorDescription: "Client unauthorized",
		HttpStatusCode:   http.StatusUnauthorized,
	}
}

func SimpleAuthServerError() *AuthError {
	return &AuthError{
		Error:            "SERVER_ERROR",
		ErrorDescription: "Internal server error",
		HttpStatusCode:   http.StatusInternalServerError,
	}
}

func SimpleAuthFailure() *AuthError {
	return &AuthError{
		Error:            "FAILURE",
		ErrorDescription: "Authentication failed",
		HttpStatusCode:   http.StatusUnauthorized,
	}
}

func NewSimpleAuthServerError(description string) *AuthError {
	return &AuthError{
		Error:            "SERVER_ERROR",
		ErrorDescription: description,
		HttpStatusCode:   http.StatusInternalServerError,
	}
}
