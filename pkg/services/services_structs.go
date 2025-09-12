package services

import (
	"time"

	"github.com/Identityplane/GoAM/pkg/model"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// LoadedRealm represents a loaded realm configuration
type LoadedRealm struct {
	Config       *model.Realm        // parsed realm config
	RealmID      string              // composite ID like "acme/customers"
	Repositories *model.Repositories // services for this realm
}

// FlowLintError represents a flow validation error
type FlowLintError struct {
	StartLineNumber int    `json:"startLineNumber"`
	StartColumn     int    `json:"startColumn"`
	EndLineNumber   int    `json:"endLineNumber"`
	EndColumn       int    `json:"endColumn"`
	Message         string `json:"message"`
	Severity        int    `json:"severity"`
}

// TimeProvider interface for time operations (useful for testing)
type TimeProvider interface {
	Now() time.Time
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	Ratio     float64
	Hits      uint64
	Misses    uint64
	KeysAdded uint64
}

// AuthzEntitlement represents an authorization entitlement
type AuthzEntitlement struct {
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Effect      string `json:"effect"`
}
