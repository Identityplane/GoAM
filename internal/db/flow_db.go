package db

import (
	"context"

	"github.com/gianlucafrei/GoAM/internal/model"
)

// FlowDB interface for flow database operations
type FlowDB interface {
	// CreateFlow creates a new flow
	CreateFlow(ctx context.Context, flow model.Flow) error

	// GetFlow retrieves a flow by its tenant, realm and id
	GetFlow(ctx context.Context, tenant, realm, id string) (*model.Flow, error)

	// GetFlowByRoute retrieves a flow by its tenant, realm and route
	GetFlowByRoute(ctx context.Context, tenant, realm, route string) (*model.Flow, error)

	// UpdateFlow updates an existing flow
	UpdateFlow(ctx context.Context, flow *model.Flow) error

	// ListFlows lists all flows for a tenant and realm
	ListFlows(ctx context.Context, tenant, realm string) ([]model.Flow, error)

	// ListAllFlows lists all flows across all tenants and realms
	ListAllFlows(ctx context.Context) ([]model.Flow, error)

	// DeleteFlow deletes a flow
	DeleteFlow(ctx context.Context, tenant, realm, id string) error
}
