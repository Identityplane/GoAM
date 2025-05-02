package sqlite_adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"goiam/internal/model"
	"time"
)

// SQLiteApplicationDB implements the ApplicationDB interface using SQLite
type SQLiteApplicationDB struct {
	db *sql.DB
}

// NewApplicationDB creates a new SQLiteApplicationDB instance
func NewApplicationDB(db *sql.DB) (*SQLiteApplicationDB, error) {
	// Check if the connection works and applications table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM applications LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if applications table exists: %w", err)
	}

	return &SQLiteApplicationDB{db: db}, nil
}

func (s *SQLiteApplicationDB) CreateApplication(ctx context.Context, app model.Application) error {
	now := time.Now()

	scopesJSON, err := json.Marshal(app.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	flowsJSON, err := json.Marshal(app.AllowedFlows)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed flows: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO applications (
			tenant, realm, client_id, client_secret, confidential, consent_required,
			description, allowed_scopes, allowed_flows, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		app.Tenant,
		app.Realm,
		app.ClientId,
		app.ClientSecret,
		app.Confidential,
		app.ConsentRequired,
		app.Description,
		scopesJSON,
		flowsJSON,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	return nil
}

func (s *SQLiteApplicationDB) GetApplication(ctx context.Context, tenant, realm, id string) (*model.Application, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_flows, created_at, updated_at
		FROM applications 
		WHERE tenant = ? AND realm = ? AND client_id = ?
	`, tenant, realm, id)

	var app model.Application
	var scopesJSON, flowsJSON string
	var createdAt, updatedAt string

	err := row.Scan(
		&app.Tenant,
		&app.Realm,
		&app.ClientId,
		&app.ClientSecret,
		&app.Confidential,
		&app.ConsentRequired,
		&app.Description,
		&scopesJSON,
		&flowsJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Parse JSON arrays
	if err := json.Unmarshal([]byte(scopesJSON), &app.AllowedScopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
	}
	if err := json.Unmarshal([]byte(flowsJSON), &app.AllowedFlows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed flows: %w", err)
	}

	// Parse timestamps
	app.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	app.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &app, nil
}

func (s *SQLiteApplicationDB) UpdateApplication(ctx context.Context, app *model.Application) error {
	now := time.Now()

	scopesJSON, err := json.Marshal(app.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	flowsJSON, err := json.Marshal(app.AllowedFlows)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed flows: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE applications SET
			client_secret = ?,
			confidential = ?,
			consent_required = ?,
			description = ?,
			allowed_scopes = ?,
			allowed_flows = ?,
			updated_at = ?
		WHERE tenant = ? AND realm = ? AND client_id = ?
	`,
		app.ClientSecret,
		app.Confidential,
		app.ConsentRequired,
		app.Description,
		scopesJSON,
		flowsJSON,
		now.Format(time.RFC3339),
		app.Tenant,
		app.Realm,
		app.ClientId,
	)
	return err
}

func (s *SQLiteApplicationDB) ListApplications(ctx context.Context, tenant, realm string) ([]model.Application, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_flows, created_at, updated_at
		FROM applications 
		WHERE tenant = ? AND realm = ?
	`, tenant, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		var scopesJSON, flowsJSON string
		var createdAt, updatedAt string

		err := rows.Scan(
			&app.Tenant,
			&app.Realm,
			&app.ClientId,
			&app.ClientSecret,
			&app.Confidential,
			&app.ConsentRequired,
			&app.Description,
			&scopesJSON,
			&flowsJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON arrays
		if err := json.Unmarshal([]byte(scopesJSON), &app.AllowedScopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
		}
		if err := json.Unmarshal([]byte(flowsJSON), &app.AllowedFlows); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed flows: %w", err)
		}

		// Parse timestamps
		app.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		app.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

func (s *SQLiteApplicationDB) ListAllApplications(ctx context.Context) ([]model.Application, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_flows, created_at, updated_at
		FROM applications
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		var scopesJSON, flowsJSON string
		var createdAt, updatedAt string

		err := rows.Scan(
			&app.Tenant,
			&app.Realm,
			&app.ClientId,
			&app.ClientSecret,
			&app.Confidential,
			&app.ConsentRequired,
			&app.Description,
			&scopesJSON,
			&flowsJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON arrays
		if err := json.Unmarshal([]byte(scopesJSON), &app.AllowedScopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
		}
		if err := json.Unmarshal([]byte(flowsJSON), &app.AllowedFlows); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed flows: %w", err)
		}

		// Parse timestamps
		app.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		app.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

func (s *SQLiteApplicationDB) DeleteApplication(ctx context.Context, tenant, realm, id string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM applications 
		WHERE tenant = ? AND realm = ? AND client_id = ?
	`, tenant, realm, id)
	return err
}
