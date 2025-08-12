package postgres_adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresFlowDB implements the FlowDB interface using PostgreSQL
type PostgresFlowDB struct {
	db *pgxpool.Pool
}

// NewPostgresFlowDB creates a new PostgresFlowDB instance
func NewPostgresFlowDB(db *pgxpool.Pool) (*PostgresFlowDB, error) {
	// Check if the connection works and flows table exists by executing a query
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM flows LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if flows table exists: %w", err)
	}

	return &PostgresFlowDB{db: db}, nil
}

func (p *PostgresFlowDB) CreateFlow(ctx context.Context, flow model.Flow) error {
	now := time.Now()

	_, err := p.db.Exec(ctx, `
		INSERT INTO flows (
			tenant, realm, id, route, active, debug_allowed, definition_yaml,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		flow.Tenant,
		flow.Realm,
		flow.Id,
		flow.Route,
		flow.Active,
		flow.DebugAllowed,
		flow.DefinitionYaml,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}
	return nil
}

func (p *PostgresFlowDB) GetFlow(ctx context.Context, tenant, realm, id string) (*model.Flow, error) {
	rows, err := p.db.Query(ctx, `
		SELECT * FROM flows 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`, tenant, realm, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flow, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.Flow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	return flow, nil
}

func (p *PostgresFlowDB) GetFlowByRoute(ctx context.Context, tenant, realm, route string) (*model.Flow, error) {
	rows, err := p.db.Query(ctx, `
		SELECT * FROM flows 
		WHERE tenant = $1 AND realm = $2 AND route = $3
	`, tenant, realm, route)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flow, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.Flow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	return flow, nil
}

func (p *PostgresFlowDB) UpdateFlow(ctx context.Context, flow *model.Flow) error {
	now := time.Now()

	_, err := p.db.Exec(ctx, `
		UPDATE flows SET
			route = $1,
			active = $2,
			debug_allowed = $3,
			definition_yaml = $4,
			updated_at = $5
		WHERE tenant = $6 AND realm = $7 AND id = $8
	`,
		flow.Route,
		flow.Active,
		flow.DebugAllowed,
		flow.DefinitionYaml,
		now,
		flow.Tenant,
		flow.Realm,
		flow.Id,
	)
	return err
}

func (p *PostgresFlowDB) ListFlows(ctx context.Context, tenant, realm string) ([]model.Flow, error) {
	rows, err := p.db.Query(ctx, `
		SELECT * FROM flows 
		WHERE tenant = $1 AND realm = $2
	`, tenant, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Flow])
	if err != nil {
		return nil, err
	}

	return flows, nil
}

func (p *PostgresFlowDB) DeleteFlow(ctx context.Context, tenant, realm, id string) error {
	_, err := p.db.Exec(ctx, `
		DELETE FROM flows 
		WHERE tenant = $1 AND realm = $2 AND id = $3
	`, tenant, realm, id)
	return err
}

func (p *PostgresFlowDB) ListAllFlows(ctx context.Context) ([]model.Flow, error) {
	rows, err := p.db.Query(ctx, `SELECT * FROM flows`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Flow])
	if err != nil {
		return nil, err
	}

	return flows, nil
}
