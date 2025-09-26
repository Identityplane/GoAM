package sqlite_adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/jmoiron/sqlx"
)

// SQLiteApplicationDB implements the ApplicationDB interface using SQLite
type SQLiteApplicationDB struct {
	db *sqlx.DB
}

// NewApplicationDB creates a new SQLiteApplicationDB instance
func NewApplicationDB(db *sql.DB) (*SQLiteApplicationDB, error) {
	// Wrap with sqlx
	sqlxDB := sqlx.NewDb(db, "sqlite3")

	// Check if the connection works and applications table exists by executing a query
	_, err := sqlxDB.Exec(`
		SELECT 1 FROM applications LIMIT 1
	`)
	if err != nil {
		log := logger.GetGoamLogger()
		log.Debug().Err(err).Msg("warning: failed to check if applications table exists")
	}

	return &SQLiteApplicationDB{db: sqlxDB}, nil
}

func (s *SQLiteApplicationDB) CreateApplication(ctx context.Context, app model.Application) error {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	// Convert slice fields to JSON
	scopesJSON, err := json.Marshal(app.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	grantsJSON, err := json.Marshal(app.AllowedGrants)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed grants: %w", err)
	}

	authFlowsJSON, err := json.Marshal(app.AllowedAuthenticationFlows)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed authentication flows: %w", err)
	}

	redirectUrisJSON, err := json.Marshal(app.RedirectUris)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect uris: %w", err)
	}

	settingsJSON, err := json.Marshal(app.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Create a temporary struct for insertion with JSON fields
	type ApplicationInsert struct {
		model.Application
		AllowedScopesJSON              string `db:"allowed_scopes"`
		AllowedGrantsJSON              string `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON string `db:"allowed_authentication_flows"`
		RedirectUrisJSON               string `db:"redirect_uris"`
		SettingsJSON                   string `db:"settings"`
	}

	insertApp := ApplicationInsert{
		Application:                    app,
		AllowedScopesJSON:              string(scopesJSON),
		AllowedGrantsJSON:              string(grantsJSON),
		AllowedAuthenticationFlowsJSON: string(authFlowsJSON),
		RedirectUrisJSON:               string(redirectUrisJSON),
		SettingsJSON:                   string(settingsJSON),
	}

	query := `
		INSERT INTO applications (
			tenant, realm, client_id, client_secret, confidential, consent_required,
			description, allowed_scopes, allowed_grants, allowed_authentication_flows,
			access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			access_token_type, access_token_algorithm, access_token_mapping,
			id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		) VALUES (
			:tenant, :realm, :client_id, :client_secret, :confidential, :consent_required,
			:description, :allowed_scopes, :allowed_grants, :allowed_authentication_flows,
			:access_token_lifetime, :refresh_token_lifetime, :id_token_lifetime,
			:access_token_type, :access_token_algorithm, :access_token_mapping,
			:id_token_algorithm, :id_token_mapping, :redirect_uris, :settings, :created_at, :updated_at
		)
	`

	_, err = s.db.NamedExecContext(ctx, query, insertApp)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	return nil
}

func (s *SQLiteApplicationDB) GetApplication(ctx context.Context, tenant, realm, id string) (*model.Application, error) {
	log := logger.GetGoamLogger()

	log.Debug().
		Str("id", id).
		Str("tenant", tenant).
		Str("realm", realm).
		Msg("sql query application")

	// Create a struct to hold the raw data from the database
	type ApplicationSelect struct {
		model.Application
		AllowedScopesJSON              string `db:"allowed_scopes"`
		AllowedGrantsJSON              string `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON string `db:"allowed_authentication_flows"`
		RedirectUrisJSON               string `db:"redirect_uris"`
		SettingsJSON                   string `db:"settings"`
	}

	var appSelect ApplicationSelect
	query := `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		FROM applications 
		WHERE tenant = ? AND realm = ? AND client_id = ?
	`

	err := s.db.GetContext(ctx, &appSelect, query, tenant, realm, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Parse JSON arrays
	if err := json.Unmarshal([]byte(appSelect.AllowedScopesJSON), &appSelect.AllowedScopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
	}
	if err := json.Unmarshal([]byte(appSelect.AllowedGrantsJSON), &appSelect.AllowedGrants); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed grants: %w", err)
	}
	if err := json.Unmarshal([]byte(appSelect.AllowedAuthenticationFlowsJSON), &appSelect.AllowedAuthenticationFlows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed authentication flows: %w", err)
	}
	if err := json.Unmarshal([]byte(appSelect.RedirectUrisJSON), &appSelect.RedirectUris); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect uris: %w", err)
	}
	if err := json.Unmarshal([]byte(appSelect.SettingsJSON), &appSelect.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &appSelect.Application, nil
}

func (s *SQLiteApplicationDB) UpdateApplication(ctx context.Context, app *model.Application) error {
	now := time.Now()
	app.UpdatedAt = now

	// Convert slice fields to JSON
	scopesJSON, err := json.Marshal(app.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	grantsJSON, err := json.Marshal(app.AllowedGrants)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed grants: %w", err)
	}

	authFlowsJSON, err := json.Marshal(app.AllowedAuthenticationFlows)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed authentication flows: %w", err)
	}

	redirectUrisJSON, err := json.Marshal(app.RedirectUris)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect uris: %w", err)
	}

	settingsJSON, err := json.Marshal(app.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Create a temporary struct for update with JSON fields
	type ApplicationUpdate struct {
		model.Application
		AllowedScopesJSON              string `db:"allowed_scopes"`
		AllowedGrantsJSON              string `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON string `db:"allowed_authentication_flows"`
		RedirectUrisJSON               string `db:"redirect_uris"`
		SettingsJSON                   string `db:"settings"`
	}

	updateApp := ApplicationUpdate{
		Application:                    *app,
		AllowedScopesJSON:              string(scopesJSON),
		AllowedGrantsJSON:              string(grantsJSON),
		AllowedAuthenticationFlowsJSON: string(authFlowsJSON),
		RedirectUrisJSON:               string(redirectUrisJSON),
		SettingsJSON:                   string(settingsJSON),
	}

	query := `
		UPDATE applications SET
			client_secret = :client_secret,
			confidential = :confidential,
			consent_required = :consent_required,
			description = :description,
			allowed_scopes = :allowed_scopes,
			allowed_grants = :allowed_grants,
			allowed_authentication_flows = :allowed_authentication_flows,
			access_token_lifetime = :access_token_lifetime,
			refresh_token_lifetime = :refresh_token_lifetime,
			id_token_lifetime = :id_token_lifetime,
			access_token_type = :access_token_type,
			access_token_algorithm = :access_token_algorithm,
			access_token_mapping = :access_token_mapping,
			id_token_algorithm = :id_token_algorithm,
			id_token_mapping = :id_token_mapping,
			redirect_uris = :redirect_uris,
			settings = :settings,
			updated_at = :updated_at
		WHERE tenant = :tenant AND realm = :realm AND client_id = :client_id
	`

	result, err := s.db.NamedExecContext(ctx, query, updateApp)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

func (s *SQLiteApplicationDB) ListApplications(ctx context.Context, tenant, realm string) ([]model.Application, error) {
	// Create a struct to hold the raw data from the database
	type ApplicationSelect struct {
		model.Application
		AllowedScopesJSON              string `db:"allowed_scopes"`
		AllowedGrantsJSON              string `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON string `db:"allowed_authentication_flows"`
		RedirectUrisJSON               string `db:"redirect_uris"`
		SettingsJSON                   string `db:"settings"`
	}

	var appSelects []ApplicationSelect
	query := `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		FROM applications 
		WHERE tenant = ? AND realm = ?
	`

	err := s.db.SelectContext(ctx, &appSelects, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	var apps []model.Application
	for _, appSelect := range appSelects {
		// Parse JSON arrays
		if err := json.Unmarshal([]byte(appSelect.AllowedScopesJSON), &appSelect.AllowedScopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.AllowedGrantsJSON), &appSelect.AllowedGrants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed grants: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.AllowedAuthenticationFlowsJSON), &appSelect.AllowedAuthenticationFlows); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed authentication flows: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.RedirectUrisJSON), &appSelect.RedirectUris); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redirect uris: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.SettingsJSON), &appSelect.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		apps = append(apps, appSelect.Application)
	}

	return apps, nil
}

func (s *SQLiteApplicationDB) ListAllApplications(ctx context.Context) ([]model.Application, error) {
	// Create a struct to hold the raw data from the database
	type ApplicationSelect struct {
		model.Application
		AllowedScopesJSON              string `db:"allowed_scopes"`
		AllowedGrantsJSON              string `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON string `db:"allowed_authentication_flows"`
		RedirectUrisJSON               string `db:"redirect_uris"`
		SettingsJSON                   string `db:"settings"`
	}

	var appSelects []ApplicationSelect
	query := `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		FROM applications
	`

	err := s.db.SelectContext(ctx, &appSelects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all applications: %w", err)
	}

	var apps []model.Application
	for _, appSelect := range appSelects {
		// Parse JSON arrays
		if err := json.Unmarshal([]byte(appSelect.AllowedScopesJSON), &appSelect.AllowedScopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.AllowedGrantsJSON), &appSelect.AllowedGrants); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed grants: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.AllowedAuthenticationFlowsJSON), &appSelect.AllowedAuthenticationFlows); err != nil {
			return nil, fmt.Errorf("failed to unmarshal allowed authentication flows: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.RedirectUrisJSON), &appSelect.RedirectUris); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redirect uris: %w", err)
		}
		if err := json.Unmarshal([]byte(appSelect.SettingsJSON), &appSelect.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		apps = append(apps, appSelect.Application)
	}

	return apps, nil
}

func (s *SQLiteApplicationDB) DeleteApplication(ctx context.Context, tenant, realm, id string) error {
	query := `
		DELETE FROM applications 
		WHERE tenant = ? AND realm = ? AND client_id = ?
	`

	result, err := s.db.ExecContext(ctx, query, tenant, realm, id)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}
