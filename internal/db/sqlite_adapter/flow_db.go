package sqlite_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/jmoiron/sqlx"
)

// SQLiteFlowDB implements the FlowDB interface using SQLite
type SQLiteFlowDB struct {
	db *sqlx.DB
}

// NewFlowDB creates a new SQLiteFlowDB instance
func NewFlowDB(db *sql.DB) (*SQLiteFlowDB, error) {
	sqlxDB := sqlx.NewDb(db, "sqlite3")

	// Check if the connection works and flows table exists by executing a query
	_, err := sqlxDB.Exec(`SELECT 1 FROM flows LIMIT 1`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if flows table exists: %w", err)
	}

	return &SQLiteFlowDB{db: sqlxDB}, nil
}

func (s *SQLiteFlowDB) CreateFlow(ctx context.Context, flow model.Flow) error {
	now := time.Now()

	// Set timestamps
	flow.CreatedAt = now
	flow.UpdatedAt = now

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO flows (
			tenant, realm, id, route, active, debug_allowed, definition_yaml,
			created_at, updated_at
		) VALUES (
			:tenant, :realm, :id, :route, :active, :debug_allowed, :definition_yaml,
			:created_at, :updated_at
		)
	`, flow)

	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}
	return nil
}

func (s *SQLiteFlowDB) GetFlow(ctx context.Context, tenant, realm, id string) (*model.Flow, error) {
	var flow model.Flow

	err := s.db.GetContext(ctx, &flow, `
		SELECT * FROM flows 
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, err
	}

	return &flow, nil
}

func (s *SQLiteFlowDB) GetFlowByRoute(ctx context.Context, tenant, realm, route string) (*model.Flow, error) {
	var flow model.Flow

	err := s.db.GetContext(ctx, &flow, `
		SELECT * FROM flows 
		WHERE tenant = ? AND realm = ? AND route = ?
	`, tenant, realm, route)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, err
	}

	return &flow, nil
}

func (s *SQLiteFlowDB) UpdateFlow(ctx context.Context, flow *model.Flow) error {
	flow.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, `
		UPDATE flows SET
			route = :route,
			active = :active,
			debug_allowed = :debug_allowed,
			definition_yaml = :definition_yaml,
			updated_at = :updated_at
		WHERE tenant = :tenant AND realm = :realm AND id = :id
	`, flow)

	return err
}

func (s *SQLiteFlowDB) ListFlows(ctx context.Context, tenant, realm string) ([]model.Flow, error) {
	var flows []model.Flow

	err := s.db.SelectContext(ctx, &flows, `
		SELECT * FROM flows 
		WHERE tenant = ? AND realm = ?
	`, tenant, realm)

	return flows, err
}

func (s *SQLiteFlowDB) DeleteFlow(ctx context.Context, tenant, realm, id string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM flows 
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, id)
	return err
}

func (s *SQLiteFlowDB) ListAllFlows(ctx context.Context) ([]model.Flow, error) {
	var flows []model.Flow

	err := s.db.SelectContext(ctx, &flows, `SELECT * FROM flows`)

	return flows, err
}
