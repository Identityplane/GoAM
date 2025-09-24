package services

import (
	"context"
	"html/template"
	"io/fs"
	"time"

	"github.com/Identityplane/GoAM/internal/lib/oauth2"
	"github.com/Identityplane/GoAM/pkg/model"
)

// Services holds all service instances
type Services struct {
	UserService                UserAdminService
	UserAttributeService       UserAttributeService
	RealmService               RealmService
	FlowService                FlowService
	ApplicationService         ApplicationService
	SessionsService            SessionsService
	StaticConfigurationService StaticConfigurationService
	OAuth2Service              OAuth2Service
	SimpleAuthService          SimpleAuthService
	JWTService                 JWTService
	CacheService               CacheService
	AdminAuthzService          AdminAuthzService
	TemplatesService           TemplatesService
}

// UserAdminService defines the business logic for user operations
type UserAdminService interface {
	// List users with pagination, returns usersn, total count and users
	ListUsers(ctx context.Context, tenant, realm string, pagination PaginationParams) ([]model.User, int64, error)
	GetUserByID(ctx context.Context, tenant, realm, userID string) (*model.User, error)
	GetUserWithAttributesByID(ctx context.Context, tenant, realm, userID string) (*model.User, error)
	UpdateUserByID(ctx context.Context, tenant, realm, userID string, updateUser model.User) (*model.User, error)
	DeleteUserByID(ctx context.Context, tenant, realm, userID string) error
	// Get user statistics
	GetUserStats(ctx context.Context, tenant, realm string) (*model.UserStats, error)
	// Create a new user
	CreateUser(ctx context.Context, tenant, realm string, createUser model.User) (*model.User, error)
	// Create a new user with attributes
	CreateUserWithAttributes(ctx context.Context, tenant, realm string, user model.User) (*model.User, error)
	// Update an existing user with attributes
	UpdateUserWithAttributes(ctx context.Context, tenant, realm string, user model.User) (*model.User, error)
	// Create or update a user with attributes (upsert)
	CreateOrUpdateUserWithAttributes(ctx context.Context, tenant, realm string, user model.User) (*model.User, error)
}

// UserAttributeService defines the business logic for user attribute operations
type UserAttributeService interface {
	// List all attributes for a user
	ListUserAttributes(ctx context.Context, tenant, realm, userID string) ([]model.UserAttribute, error)
	// Get a specific attribute by ID
	GetUserAttributeByID(ctx context.Context, tenant, realm, attributeID string) (*model.UserAttribute, error)
	// Create a new attribute for a user
	CreateUserAttribute(ctx context.Context, attribute model.UserAttribute) (*model.UserAttribute, error)
	// Update an existing attribute
	UpdateUserAttribute(ctx context.Context, attribute *model.UserAttribute) error
	// Delete a specific attribute
	DeleteUserAttribute(ctx context.Context, tenant, realm, attributeID string) error
}

// RealmService defines the business logic for realm operations
type RealmService interface {
	// GetRealm returns a loaded realm configuration by its composite ID
	GetRealm(tenant, realm string) (*LoadedRealm, bool)
	// GetAllRealms returns a map of all loaded realms with realmId as index
	GetAllRealms() (map[string]*LoadedRealm, error)
	// CreateRealm creates a new realm
	CreateRealm(realm *model.Realm) error
	// UpdateRealm updates an existing realm
	UpdateRealm(realm *model.Realm) error
	// DeleteRealm deletes a realm
	DeleteRealm(tenant, realm string) error
	// Is Tenant Name Available
	IsTenantNameAvailable(tenantName string) (bool, error)
}

// FlowService defines the business logic for flow operations
type FlowService interface {

	// GetFlow returns a flow by its ID
	GetFlowById(tenant, realm, id string) (*model.Flow, bool)

	// GetFlowByPath returns a flow by its path
	GetFlowForExecution(path string, loadedRealm *LoadedRealm) (*model.Flow, bool)

	// ListFlows returns all flows
	ListFlows(tenant, realm string) ([]model.Flow, error)

	// ListAllFlows returns all flows for all realms
	ListAllFlows() ([]model.Flow, error)

	// CreateFlow creates a new flow
	CreateFlow(tenant, realm string, flow model.Flow) error

	// UpdateFlow updates an existing flow
	UpdateFlow(tenant, realm string, flow model.Flow) error

	// DeleteFlow deletes a flow by its ID
	DeleteFlow(tenant, realm, id string) error

	// ValidateFlowDefinition validates a YAML flow definition
	ValidateFlowDefinition(content string) ([]FlowLintError, error)
}

// ApplicationService defines the business logic for application operations
type ApplicationService interface {
	// GetApplication returns an application by its ID
	GetApplication(tenant, realm, clientId string) (*model.Application, bool)

	// ListApplications returns all applications for a tenant and realm
	ListApplications(tenant, realm string) ([]model.Application, error)

	// ListAllApplications returns all applications for all realms
	ListAllApplications() ([]model.Application, error)

	// CreateApplication creates a new application
	CreateApplication(tenant, realm string, app model.Application) error

	// UpdateApplication updates an existing application
	UpdateApplication(tenant, realm string, app model.Application) error

	// DeleteApplication deletes an application by its ID
	DeleteApplication(tenant, realm, clientId string) error

	// RegenerateClientSecret generates a new client secret for an application
	RegenerateClientSecret(tenant, realm, clientId string) (string, error)

	// VerifyClientSecret verifies if a client secret matches the stored hash
	VerifyClientSecret(tenant, realm, clientId, clientSecret string) (bool, error)
}

// SessionsService defines the interface for session management
type SessionsService interface {
	// SetTimeProvider sets a custom time provider for testing
	SetTimeProvider(provider TimeProvider)

	// CreateAuthSessionObject creates a new session object but does not store it
	CreateAuthSessionObject(tenant, realm, flowId, loginUri string) (*model.AuthenticationSession, string)

	// CreateOrUpdateAuthenticationSession creates or updates an authentication session
	CreateOrUpdateAuthenticationSession(ctx context.Context, tenant, realm string, session model.AuthenticationSession) error

	// GetAuthenticationSessionByID retrieves an authentication session by its ID
	GetAuthenticationSessionByID(ctx context.Context, tenant, realm, sessionID string) (*model.AuthenticationSession, bool)

	// GetAuthenticationSession retrieves an authentication session by its hash
	GetAuthenticationSession(ctx context.Context, tenant, realm, sessionIDHash string) (*model.AuthenticationSession, bool)

	// DeleteAuthenticationSession removes an authentication session
	DeleteAuthenticationSession(ctx context.Context, tenant, realm, sessionIDHash string) error

	// CreateAuthCodeSession creates a new client session with an auth code
	CreateAuthCodeSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, codeChallenge string, codeChallengeMethod string, loginSession *model.AuthenticationSession) (string, *model.ClientSession, error)

	// CreateAccessTokenSession creates a new access token session
	CreateAccessTokenSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, lifetime int) (string, *model.ClientSession, error)

	// CreateRefreshTokenSession creates a new refresh token session
	CreateRefreshTokenSession(ctx context.Context, tenant, realm, clientID, userID string, scope []string, grantType string, expiresIn int) (string, *model.ClientSession, error)

	// GetClientSessionByAccessToken retrieves a client session by its access token
	GetClientSessionByAccessToken(ctx context.Context, tenant, realm, accessToken string) (*model.ClientSession, error)

	// LoadAndDeleteAuthCodeSession retrieves a client session by auth code and deletes it
	LoadAndDeleteAuthCodeSession(ctx context.Context, tenant, realm, authCode string) (*model.ClientSession, *model.AuthenticationSession, error)

	// LoadAndDeleteRefreshTokenSession retrieves a client session by refresh token and deletes it
	LoadAndDeleteRefreshTokenSession(ctx context.Context, tenant, realm, refreshToken string) (*model.ClientSession, error)
}

type StaticConfigurationService interface {
	LoadConfigurationFromFiles(configRoot string) error
}

// JWTService defines the business logic for JWT operations
type JWTService interface {
	// LoadPublicKeys returns the JWKS for a given tenant and realm
	LoadPublicKeys(tenant, realm string) (string, error)

	// SignJWT signs a JWT token with the key for the given tenant and realm
	SignJWT(tenant, realm string, claims map[string]interface{}) (string, error)

	// GenerateKey generates a new key for a tenant/realm
	GenerateKey(tenant, realm string) error

	// RotateKey generates a new key and disables the old one
	RotateKey(tenant, realm string) error

	// getActiveSigningKey returns an active signing key for the given tenant and realm
	// This is an internal method that takes a context
	GetActiveSigningKey(ctx context.Context, tenant, realm string) (*model.SigningKey, error)
}

// CacheService defines the interface for cache operations
// This is currently in memory only but can be extended to other cache backend such as Redis
type CacheService interface {
	// Cache stores a value in the cache with the specified TTL
	Cache(key string, value interface{}, ttl time.Duration, cost int64) error

	// Get retrieves a value from the cache by its key
	Get(key string) (interface{}, bool)

	// Invalidate removes a key from the cache
	Invalidate(key string) error

	// GetMetrics returns the metrics of the cache
	GetMetrics() CacheMetrics
}

type AdminAuthzService interface {
	GetEntitlements(user *model.User) []AuthzEntitlement
	CheckAccess(user *model.User, resource, action string) (bool, string)
	GetVisibleRealms(user *model.User) (map[string]*LoadedRealm, error)
	CreateTenant(tenantSlug, tenantName string, user *model.User) error
}

// Responseible for returning the right templates for a given realm
// this should be very fast as it is called for every request
type TemplatesService interface {
	//GetErrorTemplate(tenant, realm, flowId string) (*template.Template, error)
	CreateTemplateOverride(tenant, realm, flowId, nodeName, templateString string) error
	RemoveTemplateOverride(tenant, realm, flowId, nodeName string) error
	ListTemplateOverrides() map[string]bool
	LoadTemplateOverridesFromPath(tenant, realm, templatesPath string) error
	LoadTemplateOverridesFromFS(tenant, realm string, templatesFS fs.FS, templatesPath string) error

	// GetTemplates returns the html teplate for a specific flow
	GetTemplates(tenant, realm, flowId, nodeName string) (*template.Template, error)

	// GetErrorTemplate returns the html error template for a specific flow
	GetErrorTemplate(tenant, realm, flowId string) (*template.Template, error)
}

// OAuth2Service defines the business logic for OAuth2 operations
type OAuth2Service interface {
	// ValidateOAuth2AuthorizationRequest validates the OAuth2 authorization request
	ValidateOAuth2AuthorizationRequest(oauth2request *model.AuthorizeRequest, tenant, realm string, application *model.Application, flowId string) *oauth2.OAuth2Error

	// FinishOauth2AuthorizationEndpoint completes the OAuth2 authorization endpoint flow
	FinishOauth2AuthorizationEndpoint(session *model.AuthenticationSession, tenant, realm string) (*oauth2.AuthorizationResponse, *oauth2.OAuth2Error)

	// ProcessTokenRequest processes OAuth2 token requests for various grant types
	ProcessTokenRequest(tenant, realm string, tokenRequest *oauth2.Oauth2TokenRequest, clientAuthentication *oauth2.Oauth2ClientAuthentication) (*oauth2.Oauth2TokenResponse, *oauth2.OAuth2Error)

	// IntrospectAccessToken introspects an OAuth2 access token and returns information about it
	IntrospectAccessToken(tenant, realm string, tokenIntrospectionRequest *oauth2.TokenIntrospectionRequest) (*oauth2.TokenIntrospectionResponse, *oauth2.OAuth2Error)

	// ToQueryString converts the AuthorizationResponse to a URL query string
	ToQueryString(response *oauth2.AuthorizationResponse) string

	// GetUserClaims gets the user claims for a given client session
	GetUserClaims(user model.User, scope string) (map[string]interface{}, error)

	// GetOtherJwtClaims gets the other JWT claims for a given client
	GetOtherJwtClaims(tenant, realm, client_id string) (map[string]interface{}, error)
}

// SimpleAuthService defines the business logic for Simple Auth operations
type SimpleAuthService interface {
	// VerifySimpleAuthFlowRequest validates the Simple Auth flow request
	VerifySimpleAuthFlowRequest(ctx context.Context, req *model.SimpleAuthRequest, application *model.Application, flow *model.Flow) error

	// FinishSimpleAuthFlow completes the Simple Auth flow and returns tokens
	FinishSimpleAuthFlow(ctx context.Context, session *model.AuthenticationSession, tenant, realm string) (*model.SimpleAuthResponse, *model.SimpleAuthError)
}
