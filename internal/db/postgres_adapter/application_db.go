package postgres_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresApplicationDB implements the ApplicationDB interface using PostgreSQL
type PostgresApplicationDB struct {
	db *pgxpool.Pool
}

// NewPostgresApplicationDB creates a new PostgresApplicationDB instance
func NewPostgresApplicationDB(db *pgxpool.Pool) (*PostgresApplicationDB, error) {
	// Check if the connection works and applications table exists by executing a query
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM applications LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if applications table exists: %w", err)
	}

	return &PostgresApplicationDB{db: db}, nil
}

func (p *PostgresApplicationDB) CreateApplication(ctx context.Context, app model.Application) error {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	// Convert JSON fields to strings
	allowedScopesJSON, err := json.Marshal(app.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}
	allowedGrantsJSON, err := json.Marshal(app.AllowedGrants)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed grants: %w", err)
	}
	allowedAuthFlowsJSON, err := json.Marshal(app.AllowedAuthenticationFlows)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed authentication flows: %w", err)
	}
	accessTokenMappingJSON, err := json.Marshal(app.AccessTokenMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal access token mapping: %w", err)
	}
	idTokenMappingJSON, err := json.Marshal(app.IdTokenMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal id token mapping: %w", err)
	}
	redirectUrisJSON, err := json.Marshal(app.RedirectUris)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect uris: %w", err)
	}
	settingsJSON, err := json.Marshal(app.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	_, err = p.db.Exec(ctx, `
		INSERT INTO applications (
			tenant, realm, client_id, client_secret, confidential, consent_required,
			description, allowed_scopes, allowed_grants, allowed_authentication_flows,
			access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			access_token_type, access_token_algorithm, access_token_mapping,
			id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`,
		app.Tenant,
		app.Realm,
		app.ClientId,
		app.ClientSecret,
		app.Confidential,
		app.ConsentRequired,
		app.Description,
		allowedScopesJSON,
		allowedGrantsJSON,
		allowedAuthFlowsJSON,
		app.AccessTokenLifetime,
		app.RefreshTokenLifetime,
		app.IdTokenLifetime,
		app.AccessTokenType,
		app.AccessTokenAlgorithm,
		accessTokenMappingJSON,
		app.IdTokenAlgorithm,
		idTokenMappingJSON,
		redirectUrisJSON,
		settingsJSON,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	return nil
}

func (p *PostgresApplicationDB) GetApplication(ctx context.Context, tenant, realm, id string) (*model.Application, error) {
	log := logger.GetGoamLogger()

	log.Debug().
		Str("id", id).
		Str("tenant", tenant).
		Str("realm", realm).
		Msg("sql query application")

	query := `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		FROM applications
		WHERE tenant = $1 AND realm = $2 AND client_id = $3
	`

	rows, err := p.db.Query(ctx, query, tenant, realm, id)
	if err != nil {
		return nil, fmt.Errorf("select application: %w", err)
	}
	defer rows.Close()

	// Use a temporary struct to handle JSONB fields properly
	type ApplicationRow struct {
		model.Application
		AllowedScopesJSON              []byte `db:"allowed_scopes"`
		AllowedGrantsJSON              []byte `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON []byte `db:"allowed_authentication_flows"`
		AccessTokenMappingJSON         []byte `db:"access_token_mapping"`
		IdTokenMappingJSON             []byte `db:"id_token_mapping"`
		RedirectUrisJSON               []byte `db:"redirect_uris"`
		SettingsJSON                   []byte `db:"settings"`
	}

	appRow, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[ApplicationRow])
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan application: %w", err)
	}

	// Unmarshal JSON fields
	json.Unmarshal(appRow.AllowedScopesJSON, &appRow.AllowedScopes)
	json.Unmarshal(appRow.AllowedGrantsJSON, &appRow.AllowedGrants)
	json.Unmarshal(appRow.AllowedAuthenticationFlowsJSON, &appRow.AllowedAuthenticationFlows)
	json.Unmarshal(appRow.AccessTokenMappingJSON, &appRow.AccessTokenMapping)
	json.Unmarshal(appRow.IdTokenMappingJSON, &appRow.IdTokenMapping)
	json.Unmarshal(appRow.RedirectUrisJSON, &appRow.RedirectUris)
	json.Unmarshal(appRow.SettingsJSON, &appRow.Settings)

	return &appRow.Application, nil
}

func (p *PostgresApplicationDB) UpdateApplication(ctx context.Context, app *model.Application) error {
	now := time.Now()
	app.UpdatedAt = now

	// Convert JSON fields to strings
	allowedScopesJSON, err := json.Marshal(app.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}
	allowedGrantsJSON, err := json.Marshal(app.AllowedGrants)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed grants: %w", err)
	}
	allowedAuthFlowsJSON, err := json.Marshal(app.AllowedAuthenticationFlows)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed authentication flows: %w", err)
	}
	accessTokenMappingJSON, err := json.Marshal(app.AccessTokenMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal access token mapping: %w", err)
	}
	idTokenMappingJSON, err := json.Marshal(app.IdTokenMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal id token mapping: %w", err)
	}
	redirectUrisJSON, err := json.Marshal(app.RedirectUris)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect uris: %w", err)
	}
	settingsJSON, err := json.Marshal(app.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	_, err = p.db.Exec(ctx, `
		UPDATE applications SET
			client_secret = $1,
			confidential = $2,
			consent_required = $3,
			description = $4,
			allowed_scopes = $5,
			allowed_grants = $6,
			allowed_authentication_flows = $7,
			access_token_lifetime = $8,
			refresh_token_lifetime = $9,
			id_token_lifetime = $10,
			access_token_type = $11,
			access_token_algorithm = $12,
			access_token_mapping = $13,
			id_token_algorithm = $14,
			id_token_mapping = $15,
			redirect_uris = $16,
			settings = $17,
			updated_at = $18
		WHERE tenant = $19 AND realm = $20 AND client_id = $21
	`,
		app.ClientSecret,
		app.Confidential,
		app.ConsentRequired,
		app.Description,
		allowedScopesJSON,
		allowedGrantsJSON,
		allowedAuthFlowsJSON,
		app.AccessTokenLifetime,
		app.RefreshTokenLifetime,
		app.IdTokenLifetime,
		app.AccessTokenType,
		app.AccessTokenAlgorithm,
		accessTokenMappingJSON,
		app.IdTokenAlgorithm,
		idTokenMappingJSON,
		redirectUrisJSON,
		settingsJSON,
		now,
		app.Tenant,
		app.Realm,
		app.ClientId,
	)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}
	return nil
}

func (p *PostgresApplicationDB) ListApplications(ctx context.Context, tenant, realm string) ([]model.Application, error) {
	query := `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		FROM applications
		WHERE tenant = $1 AND realm = $2
	`

	rows, err := p.db.Query(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("select applications: %w", err)
	}
	defer rows.Close()

	// Use a temporary struct to handle JSONB fields properly
	type ApplicationRow struct {
		model.Application
		AllowedScopesJSON              []byte `db:"allowed_scopes"`
		AllowedGrantsJSON              []byte `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON []byte `db:"allowed_authentication_flows"`
		AccessTokenMappingJSON         []byte `db:"access_token_mapping"`
		IdTokenMappingJSON             []byte `db:"id_token_mapping"`
		RedirectUrisJSON               []byte `db:"redirect_uris"`
		SettingsJSON                   []byte `db:"settings"`
	}

	appRows, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ApplicationRow])
	if err != nil {
		return nil, fmt.Errorf("collect applications: %w", err)
	}

	var apps []model.Application
	for _, appRow := range appRows {
		// Unmarshal JSON fields
		json.Unmarshal(appRow.AllowedScopesJSON, &appRow.AllowedScopes)
		json.Unmarshal(appRow.AllowedGrantsJSON, &appRow.AllowedGrants)
		json.Unmarshal(appRow.AllowedAuthenticationFlowsJSON, &appRow.AllowedAuthenticationFlows)
		json.Unmarshal(appRow.AccessTokenMappingJSON, &appRow.AccessTokenMapping)
		json.Unmarshal(appRow.IdTokenMappingJSON, &appRow.IdTokenMapping)
		json.Unmarshal(appRow.RedirectUrisJSON, &appRow.RedirectUris)
		json.Unmarshal(appRow.SettingsJSON, &appRow.Settings)

		apps = append(apps, appRow.Application)
	}

	return apps, nil
}

func (p *PostgresApplicationDB) ListAllApplications(ctx context.Context) ([]model.Application, error) {
	query := `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, settings, created_at, updated_at
		FROM applications
	`

	rows, err := p.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select applications: %w", err)
	}
	defer rows.Close()

	// Use a temporary struct to handle JSONB fields properly
	type ApplicationRow struct {
		model.Application
		AllowedScopesJSON              []byte `db:"allowed_scopes"`
		AllowedGrantsJSON              []byte `db:"allowed_grants"`
		AllowedAuthenticationFlowsJSON []byte `db:"allowed_authentication_flows"`
		AccessTokenMappingJSON         []byte `db:"access_token_mapping"`
		IdTokenMappingJSON             []byte `db:"id_token_mapping"`
		RedirectUrisJSON               []byte `db:"redirect_uris"`
		SettingsJSON                   []byte `db:"settings"`
	}

	appRows, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ApplicationRow])
	if err != nil {
		return nil, fmt.Errorf("collect applications: %w", err)
	}

	var apps []model.Application
	for _, appRow := range appRows {
		// Unmarshal JSON fields
		json.Unmarshal(appRow.AllowedScopesJSON, &appRow.AllowedScopes)
		json.Unmarshal(appRow.AllowedGrantsJSON, &appRow.AllowedGrants)
		json.Unmarshal(appRow.AllowedAuthenticationFlowsJSON, &appRow.AllowedAuthenticationFlows)
		json.Unmarshal(appRow.AccessTokenMappingJSON, &appRow.AccessTokenMapping)
		json.Unmarshal(appRow.IdTokenMappingJSON, &appRow.IdTokenMapping)
		json.Unmarshal(appRow.RedirectUrisJSON, &appRow.RedirectUris)
		json.Unmarshal(appRow.SettingsJSON, &appRow.Settings)

		apps = append(apps, appRow.Application)
	}

	return apps, nil
}

func (p *PostgresApplicationDB) DeleteApplication(ctx context.Context, tenant, realm, id string) error {
	query := `
		DELETE FROM applications 
		WHERE tenant = $1 AND realm = $2 AND client_id = $3
	`

	result, err := p.db.Exec(ctx, query, tenant, realm, id)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}
