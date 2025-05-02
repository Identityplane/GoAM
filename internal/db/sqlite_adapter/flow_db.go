package sqlite_adapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"goiam/internal/model"
	"time"
)

// SQLiteFlowDB implements the FlowDB interface using SQLite
type SQLiteFlowDB struct {
	db *sql.DB
}

// NewFlowDB creates a new SQLiteFlowDB instance
func NewFlowDB(db *sql.DB) (*SQLiteFlowDB, error) {
	// Check if the connection works and flows table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM flows LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if flows table exists: %w", err)
	}

	return &SQLiteFlowDB{db: db}, nil
}

func (s *SQLiteFlowDB) CreateFlow(ctx context.Context, flow model.Flow) error {
	now := time.Now()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO flows (
			tenant, realm, id, route, active, definition_yaml,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		flow.Tenant,
		flow.Realm,
		flow.Id,
		flow.Route,
		flow.Active,
		flow.DefintionYaml,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}
	return nil
}

func (s *SQLiteFlowDB) GetFlow(ctx context.Context, tenant, realm, id string) (*model.Flow, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT tenant, realm, id, route, active, definition_yaml,
		       created_at, updated_at
		FROM flows 
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, id)

	var flow model.Flow
	var createdAt, updatedAt string

	err := row.Scan(
		&flow.Tenant,
		&flow.Realm,
		&flow.Id,
		&flow.Route,
		&flow.Active,
		&flow.DefintionYaml,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Parse timestamps
	flow.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	flow.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &flow, nil
}

func (s *SQLiteFlowDB) GetFlowByRoute(ctx context.Context, tenant, realm, route string) (*model.Flow, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT tenant, realm, id, route, active, definition_yaml,
		       created_at, updated_at
		FROM flows 
		WHERE tenant = ? AND realm = ? AND route = ?
	`, tenant, realm, route)

	var flow model.Flow
	var createdAt, updatedAt string

	err := row.Scan(
		&flow.Tenant,
		&flow.Realm,
		&flow.Id,
		&flow.Route,
		&flow.Active,
		&flow.DefintionYaml,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Parse timestamps
	flow.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	flow.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &flow, nil
}

func (s *SQLiteFlowDB) UpdateFlow(ctx context.Context, flow *model.Flow) error {
	now := time.Now()

	_, err := s.db.ExecContext(ctx, `
		UPDATE flows SET
			route = ?,
			active = ?,
			definition_yaml = ?,
			updated_at = ?
		WHERE tenant = ? AND realm = ? AND id = ?
	`,
		flow.Route,
		flow.Active,
		flow.DefintionYaml,
		now.Format(time.RFC3339),
		flow.Tenant,
		flow.Realm,
		flow.Id,
	)
	return err
}

func (s *SQLiteFlowDB) ListFlows(ctx context.Context, tenant, realm string) ([]model.Flow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tenant, realm, id, route, active, definition_yaml,
		       created_at, updated_at
		FROM flows 
		WHERE tenant = ? AND realm = ?
	`, tenant, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []model.Flow
	for rows.Next() {
		var flow model.Flow
		var createdAt, updatedAt string

		err := rows.Scan(
			&flow.Tenant,
			&flow.Realm,
			&flow.Id,
			&flow.Route,
			&flow.Active,
			&flow.DefintionYaml,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		flow.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		flow.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return flows, nil
}

func (s *SQLiteFlowDB) DeleteFlow(ctx context.Context, tenant, realm, id string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM flows 
		WHERE tenant = ? AND realm = ? AND id = ?
	`, tenant, realm, id)
	return err
}

func (s *SQLiteFlowDB) ListAllFlows(ctx context.Context) ([]model.Flow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tenant, realm, id, route, active, definition_yaml,
		       created_at, updated_at
		FROM flows
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []model.Flow
	for rows.Next() {
		var flow model.Flow
		var createdAt, updatedAt string

		err := rows.Scan(
			&flow.Tenant,
			&flow.Realm,
			&flow.Id,
			&flow.Route,
			&flow.Active,
			&flow.DefintionYaml,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		flow.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		flow.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return flows, nil
}
