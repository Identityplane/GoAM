# JSON Authentication API

## Overview

The JSON Authentication API is designed to integrate with native mobile applications and other first party clients. This API allows applications to directly interact with login, registration, and management flows without requiring web-based redirects or OAuth2 flows.

#### First Party Application:
An application that the owner of the realm directly controls. This could be the mobile app or web page. In that case the simple flow can be used but it should be considered if standard OAuth2 offers better decoupling.

#### Third Party Applications
Any application that is controlled by somebody else than the owner of the realm should not use this API, but the OAuth 2.1 interface.



## Core Concepts

### Purpose
- **Native Mobile Apps**: Login on 1. party mobile app applications that display a native login screen.
- **User Management** User management flows like change password, add email, add device etc can be executed through this api.
- **Future GoAM Internal Use**: In the future we might use this api internally. e.g. to replace the html renderer with a SPA that uses this api.

### Key Principles
1. **Session-based Flow Management**: Each flow starts with a GET request to obtain session credentials
2. **Application Integration**: Uses `client_id` for OAuth2 token generation and flow authorization

## API Schema (Example Flow)

### Base URL Structure
```
GET/POST /{tenant}/{realm}/api/v1/{flow_route}
```

### Request/Response Flow

#### 1. Initial Flow Request (GET)
```http
GET /acme/customers/api/v1/username-password-register?client_id=customers-app
Accept: application/json
```

**Response:**
```json
{
  "executionId": "57bf784b-8af9-4049-b3b5-9acaa2462683",
  "sessionId": "4fd68cbe-4162-4066-ad4f-8440ef080710",
  "currentNode": "askUsername",
  "prompts": {
    "username": "text"
  }
}
```

#### 2. Flow Continuation (POST)
```http
POST /acme/customers/api/v1/username-password-register
Content-Type: application/json

{
  "executionId": "57bf784b-8af9-4049-b3b5-9acaa2462683",
  "sessionId": "4fd68cbe-4162-4066-ad4f-8440ef080710",
  "currentNode": "askUsername",
  "responses": {
    "username": "testuser"
  }
}
```

**Response:**
```json
{
  "executionId": "57bf784b-8af9-4049-b3b5-9acaa2462683",
  "sessionId": "4fd68cbe-4162-4066-ad4f-8440ef080710",
  "currentNode": "askPassword",
  "prompts": {
    "password": "password"
  }
}
```

#### 3. Flow Completion (POST)
```http
POST /acme/customers/api/v1/username-password-register
Content-Type: application/json

{
  "executionId": "57bf784b-8af9-4049-b3b5-9acaa2462683",
  "sessionId": "4fd68cbe-4162-4066-ad4f-8440ef080710",
  "currentNode": "askPassword",
  "responses": {
    "password": "testuser"
  }
}
```

**Success Response:**
```json
{
  "executionId": "57bf784b-8af9-4049-b3b5-9acaa2462683",
  "sessionId": "4fd68cbe-4162-4066-ad4f-8440ef080710",
  "currentNode": "registerSuccess",
  "result": {
    "success": true,
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "refresh_token": "refresh_token_here",
    "expires_in": 3600,
    "refresh_token_expires_in": 31536000,
    "user": {
      "sub": "user_id_here"
    }
  }
}
```

**Failure Response:**
```json
{
  "executionId": "57bf784b-8af9-4049-b3b5-9acaa2462683",
  "sessionId": "4fd68cbe-4162-4066-ad4f-8440ef080710",
  "currentNode": "failureResult",
  "result": {
    "success": false,
    "message": "Authentication failed"
  }
}
```


## Application Integration

### Prerequisites

#### 1. Application Registration
The application must be registered in the realm configuration:

```yaml
applications:
  customers-app:
    client_id: customers-app
    client_secret: customers-app-secret
    allowed_grants:
      - simple-body          # Required for JSON API
      - simple-cookie
      - refresh_token
    redirect_uris:
      - http://localhost:3000
    allowed_authentication_flows:
      - "*"                  # Allow all flows
    access_token_lifetime: 3600
    refresh_token_lifetime: 31536000
    id_token_lifetime: 3600
```

#### 2. Flow Configuration
The desired flow must be enabled in the realm:

```yaml
flows:
  username-password-register:
    route: /username-password-register
    active: yes
    definition_location: username-password-register.yaml
```

### Integration Steps

#### 1. Initialize Flow
```javascript
// Start authentication flow
const response = await fetch('/acme/customers/api/v1/username-password-register?client_id=customers-app', {
  method: 'GET',
  headers: {
    'Accept': 'application/json'
  }
});

const flowData = await response.json();
// Store executionId and sessionId for subsequent requests
```

#### 2. Process Flow Steps
```javascript
// Continue flow with user input
const response = await fetch('/acme/customers/api/v1/username-password-register', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    executionId: flowData.executionId,
    sessionId: flowData.sessionId,
    currentNode: flowData.currentNode,
    responses: {
      username: userInput
    }
  })
});

const nextStep = await response.json();
```

#### 3. Handle Completion
```javascript
// Check if flow is complete
if (nextStep.result) {
  if (nextStep.result.success) {
    // Authentication successful
    const accessToken = nextStep.result.access_token;
    const refreshToken = nextStep.result.refresh_token;
    // Store tokens and proceed to authenticated state
  } else {
    // Authentication failed
    console.error('Authentication failed:', nextStep.result.message);
  }
} else {
  // Continue with next step
  // Present prompts to user and collect responses
}
```

## Advanced Configuration

### Query Parameters

#### client_id (Required for Token Generation)
- **Purpose**: Identifies the application for OAuth2 token generation
- **Usage**: `?client_id=customers-app`
- **Effect**: Enables OAuth2 token generation in successful flows

#### scope (Optional)
- **Purpose**: Specifies requested permissions
- **Usage**: `?scope=read write`
- **Effect**: Included in generated access tokens

#### response_type (Optional)
- **Purpose**: Controls token generation behavior
- **Usage**: `?response_type=refresh_token`
- **Effect**: Determines if refresh tokens are issued


## Security Considerations

### Session Management
- **Session IDs**: Sensitive, used for authentication
- **Execution IDs**: Non-sensitive, used for debugging
- **Session Invalidation**: Sessions are invalidated after completion or error. Abandoned sessions expire after the login timeout.


## Error Handling

### HTTP Status Codes
- **200 OK**: Flow step completed successfully
- **400 Bad Request**: Invalid request (missing client_id, malformed JSON)
- **404 Not Found**: Flow or realm not found
- **500 Internal Server Error**: Server error

### Error Response Format
```json
{
  "error": {
    "error": "INVALID_REQUEST",
    "error_description": "Client ID is required"
  }
}
```

### Common Error Scenarios
1. **Missing client_id**: Required for OAuth2 token generation
2. **Invalid session**: Session ID expired or invalid
3. **Flow not found**: Requested flow doesn't exist or is inactive
4. **Malformed request**: Invalid JSON or missing required fields
5. Flow internal errors
