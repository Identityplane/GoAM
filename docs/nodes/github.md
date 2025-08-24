# GitHub OAuth Nodes

The GitHub nodes provide OAuth 2.0 authentication functionality for GitHub integration. These nodes allow users to authenticate using their GitHub accounts and create user profiles based on GitHub user data.

## Overview

The GitHub OAuth implementation consists of two main nodes:
- **githubLogin**: Handles the OAuth flow, including redirect to GitHub and processing the authorization code
- **githubCreateUser**: Creates a new user account using information from GitHub OAuth authentication

## GitHub Attribute

Both nodes work with a user attribute of type `"identityplane:github"` that contains:

- **GitHubUserID** (`string`): The unique GitHub user ID
- **GitHubUsername** (`string`): The GitHub username/login
- **GitHubEmail** (`string`): The user's GitHub email address
- **GitHubAvatarURL** (`string`): URL to the user's GitHub avatar image
- **GitHubAccessToken** (`string`): OAuth access token for GitHub API calls
- **GitHubRefreshToken** (`string`): OAuth refresh token for renewing access
- **GitHubTokenType** (`string`): Type of OAuth token (typically "bearer")
- **GitHubScope** (`string`): OAuth scopes granted to the application

## githubLogin Node

### Purpose
Handles the complete GitHub OAuth authentication flow, including redirect to GitHub, token exchange, and user lookup.

### Node Type
`NodeTypeQueryWithLogic`

### Required Context
- No specific context required (handles both initial redirect and callback)

### Custom Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `github-client-id` | string | Yes | The client ID of the GitHub OAuth application |
| `github-client-secret` | string | Yes | The client secret of the GitHub OAuth application |
| `github-scope` | string | Yes | Comma-separated list of OAuth scopes to request from GitHub |
| `create-user-if-not-exists` | string | No | If "true", indicates that new users should be created if they don't exist |

### Behavior

1. **Initial Call** (no `code` input):
   - Generates a GitHub OAuth authorization URL
   - Redirects the user to GitHub for authentication
   - Returns a `__redirect` prompt with the GitHub authorization URL

2. **Callback Call** (with `code` input):
   - Exchanges the authorization code for an access token
   - Retrieves user data from GitHub using the access token
   - Creates a `GitHubAttributeValue` and stores it in the session context
   - Checks if a user with this GitHub ID already exists
   - Returns appropriate result state based on user existence

### Output Context
- `github`: JSON-serialized GitHub attribute value containing all user data
- `github-username`: GitHub username
- `github-access-token`: OAuth access token
- `github-refresh-token`: OAuth refresh token
- `github-token-type`: Token type
- `github-scope`: Granted OAuth scopes
- `github-user-id`: GitHub user ID
- `github-avatar-url`: Avatar image URL
- `github-email`: User's email address

### Result States
- `existing-user`: User with this GitHub ID already exists in the system
- `new-user`: This is a new GitHub user (no existing account found)
- `failure`: OAuth flow failed (invalid token, API error, etc.)

## githubCreateUser Node

### Purpose
Creates a new user account using information from GitHub OAuth authentication and persists it to the database.

### Node Type
`NodeTypeLogic`

### Required Context
- `github`: GitHub attribute value (set by githubLogin node)

### Custom Configuration Options
No custom configuration options required.

### Behavior

1. **User Creation**:
   - Parses the GitHub attribute from the session context
   - Creates a new user with "active" status if none exists
   - Adds the GitHub attribute to the user with the GitHub user ID as the index
   - Persists the user to the database using `CreateOrUpdate`

2. **Attribute Management**:
   - Automatically creates a `UserAttribute` of type `"identityplane:github"`
   - Uses the GitHub user ID as the attribute index for efficient lookups
   - Stores the complete `GitHubAttributeValue` structure

### Output Context
- `user`: The created or updated user object

### Result States
- `created`: User successfully created/updated with GitHub attributes

## OAuth Flow Integration

The GitHub nodes work together to provide a complete OAuth authentication flow:

1. **Initial Request**: `githubLogin` node redirects user to GitHub
2. **GitHub Authentication**: User authenticates with GitHub and grants permissions
3. **Callback Processing**: `githubLogin` processes the authorization code and retrieves user data
4. **User Creation**: `githubCreateUser` creates the user account and persists GitHub attributes
5. **Session Management**: User is now authenticated and can access the system

## Security Features

- **OAuth 2.0 Compliance**: Follows OAuth 2.0 standards for secure authentication
- **Token Management**: Handles access tokens and refresh tokens securely
- **Scope Control**: Configurable OAuth scopes limit application permissions
- **User Verification**: Validates GitHub user data before account creation
- **Database Persistence**: Securely stores OAuth tokens and user attributes

## Usage Examples

### Basic GitHub OAuth Flow
```yaml
# GitHub login node configuration
githubLogin:
  type: githubLogin
  customConfig:
    github-client-id: "your_github_client_id"
    github-client-secret: "your_github_client_secret"
    github-scope: "user:email"
    create-user-if-not-exists: "true"

# GitHub user creation node configuration
githubCreateUser:
  type: githubCreateUser
```

### Complete Authentication Flow
```yaml
# Flow configuration example
flow:
  - githubLogin
  - githubCreateUser
  - success
```

## Integration Notes

- **GitHub App Setup**: Requires a GitHub OAuth application with proper client ID and secret. For detailed setup instructions, see [Creating an OAuth App](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app).
- **Redirect URI**: Must be configured in GitHub OAuth app settings
- **Scope Selection**: Choose appropriate scopes based on required user data
- **User Management**: New users are automatically created with "active" status
- **Attribute Indexing**: GitHub user ID is used as the attribute index for efficient lookups
- **Database Schema**: Requires user attributes table with JSON value support
- **Error Handling**: Comprehensive error handling for OAuth failures and API errors
