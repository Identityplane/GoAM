# OAuth 2.1 API

## Overview

GoAM implements OAuth 2.1 (RFC 6749) and OpenID Connect 1.0 for third-party application authentication. This API is designed for applications controlled by external parties and provides standard OAuth2 flows with additional flow-based authentication.

## Endpoints Overview

- **Authorization Endpoint** - `GET /{tenant}/{realm}/oauth2/authorize` - Initiates OAuth2 flow with optional flow selection
- **Token Endpoint** - `POST /{tenant}/{realm}/oauth2/token` - Exchanges codes for tokens and handles refresh
- **Token Introspection** - `POST /{tenant}/{realm}/oauth2/introspect` - RFC 7662 token introspection
- **OpenID Connect Discovery** - `GET /{tenant}/{realm}/oauth2/.well-known/openid-configuration` - OIDC metadata
- **JWKs Endpoint** - `GET /{tenant}/{realm}/oauth2/.well-known/jwks.json` - JSON Web Key Set
- **UserInfo Endpoint** - `GET /{tenant}/{realm}/oauth2/userinfo` - User information for access tokens

## Standards Compliance

- **OAuth 2.1**: with security best practices
- **OpenID Connect 1.0**: OIDC specification
- **PKCE**: RFC 7636 for SPA security
- **Token Introspection**: RFC 7662
- **JWK**: RFC 7517 for key management

## Base URL Structure

```
/{tenant}/{realm}/oauth2/{endpoint}
```

## Implemented Endpoints

### 1. Authorization Endpoint
**GET** `/{tenant}/{realm}/oauth2/authorize`

Initiates the OAuth2 authorization flow with optional flow selection.

#### Parameters
- `client_id` (required): Application identifier
- `redirect_uri` (required): Registered redirect URI
- `response_type` (required): Must be "code"
- `scope` (optional): Space-separated scopes
- `state` (optional): CSRF protection parameter
- `code_challenge` (required): PKCE code challenge
- `code_challenge_method` (required): PKCE method (S256)
- `flow` (optional): Specific authentication flow to use
- `prompt` (optional): OIDC prompt parameter (login, none)
- `nonce` (optional): OIDC nonce parameter

#### Flow Parameter
The `flow` parameter allows selecting a specific authentication flow:
- If omitted, uses the first flow from `allowed_authentication_flows` in application config
- If specified, must be in the application's `allowed_authentication_flows` list
- Example: `?flow=username-password-login`

#### Response
- **302 Found**: Redirects to authentication flow or client redirect URI
- **400 Bad Request**: Invalid parameters

### 2. Token Endpoint
**POST** `/{tenant}/{realm}/oauth2/token`

Exchanges authorization codes for access tokens and handles token refresh.

#### Content-Type
`application/x-www-form-urlencoded`

#### Parameters
- `grant_type` (required): "authorization_code", "refresh_token", or "client_credentials"
- `code` (required for authorization_code): Authorization code from authorize endpoint
- `redirect_uri` (required for authorization_code): Must match authorize request
- `code_verifier` (required for authorization_code): PKCE code verifier
- `refresh_token` (required for refresh_token): Valid refresh token
- `scope` (optional): Requested scopes
- `client_id` (required): Application identifier
- `client_secret` (required): Application secret

#### Client Authentication
Supports two methods (Basic Auth takes priority):
1. **Basic Authentication**: `Authorization: Basic {base64(client_id:client_secret)}`
2. **Form Parameters**: `client_id` and `client_secret` in request body

#### Response
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_here",
  "scope": "openid profile"
}
```

### 3. Token Introspection Endpoint
**POST** `/{tenant}/{realm}/oauth2/introspect`

Introspects access tokens according to RFC 7662.

#### Content-Type
`application/x-www-form-urlencoded`

#### Parameters
- `token` (required): Access token to introspect
- `token_type_hint` (optional): Hint about token type (not implemented)

#### Response
```json
{
  "active": true,
  "scope": "openid profile",
  "client_id": "example-client",
  "username": "user123",
  "exp": 1640995200
}
```

### 4. OpenID Connect Discovery
**GET** `/{tenant}/{realm}/oauth2/.well-known/openid-configuration`

Returns OpenID Connect configuration metadata.

#### Response
```json
{
  "issuer": "https://example.com/acme/customers",
  "authorization_endpoint": "https://example.com/acme/customers/oauth2/authorize",
  "token_endpoint": "https://example.com/acme/customers/oauth2/token",
  "userinfo_endpoint": "https://example.com/acme/customers/oauth2/userinfo",
  "jwks_uri": "https://example.com/acme/customers/oauth2/.well-known/jwks.json",
  "scopes_supported": ["openid", "profile"],
  "response_types_supported": ["code"],
  "response_modes_supported": ["query"],
  "grant_types_supported": ["authorization_code", "refresh_token", "client_credentials"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["ES256"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic"],
  "claims_supported": ["sub", "iss", "aud", "exp", "iat", "auth_time", "nonce", "acr", "amr", "name", "given_name", "family_name", "username"]
}
```

### 5. JWKs Endpoint
**GET** `/{tenant}/{realm}/oauth2/.well-known/jwks.json`

Returns JSON Web Key Set for token verification.

#### Response
```json
{
  "keys": [
    {
      "kty": "EC",
      "use": "sig",
      "crv": "P-256",
      "kid": "key-id",
      "x": "base64-encoded-x-coordinate",
      "y": "base64-encoded-y-coordinate"
    }
  ]
}
```

### 6. UserInfo Endpoint
**GET** `/{tenant}/{realm}/oauth2/userinfo`

Returns user information for authenticated access tokens.

#### Headers
- `Authorization: Bearer {access_token}` (required)

#### Response
```json
{
  "sub": "user123",
  "name": "John Doe",
  "given_name": "John",
  "family_name": "Doe",
  "username": "johndoe",
  "email": "john@example.com"
}
```

## Supported Grant Types

### 1. Authorization Code Flow (PKCE)
Standard OAuth2 authorization code flow with PKCE for security.

**Flow:**
1. Client redirects user to authorization endpoint
2. User authenticates via selected flow
3. Server redirects to client with authorization code
4. Client exchanges code for tokens at token endpoint

### 2. Refresh Token
Refresh expired access tokens using refresh tokens.

**Flow:**
1. Client requests new access token using refresh token
2. Server validates refresh token and issues new access token

### 3. Client Credentials
Server-to-server authentication for machine-to-machine communication.

**Flow:**
1. Client authenticates with client credentials
2. Server issues access token directly

## Application Configuration

Applications must be registered in the realm configuration:

```yaml
applications:
  third-party-app:
    client_id: third-party-app
    client_secret: secret123
    allowed_grants:
      - authorization_code
      - refresh_token
      - client_credentials
    redirect_uris:
      - https://app.example.com/callback
    allowed_authentication_flows:
      - username-password-login
      - email-password-login
    access_token_lifetime: 3600
    refresh_token_lifetime: 31536000
```

## Error Responses

All endpoints return standard OAuth2 error responses:

```json
{
  "error": "invalid_request",
  "error_description": "Missing required parameter: client_id"
}
```

Common error codes:
- `invalid_request`: Missing or invalid parameters
- `unauthorized_client`: Invalid client credentials
- `invalid_scope`: Requested scope not allowed
- `server_error`: Internal server error

