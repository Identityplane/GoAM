package postgres_adapter

import (
	"context"
	"encoding/json"
	"errors"
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

	// Convert JSON fields to strings
	allowedScopesJSON, _ := json.Marshal(app.AllowedScopes)
	allowedGrantsJSON, _ := json.Marshal(app.AllowedGrants)
	allowedAuthFlowsJSON, _ := json.Marshal(app.AllowedAuthenticationFlows)
	accessTokenMappingJSON, _ := json.Marshal(app.AccessTokenMapping)
	idTokenMappingJSON, _ := json.Marshal(app.IdTokenMapping)
	redirectUrisJSON, _ := json.Marshal(app.RedirectUris)

	_, err := p.db.Exec(ctx, `
		INSERT INTO applications (
			tenant, realm, client_id, client_secret, confidential, consent_required,
			description, allowed_scopes, allowed_grants, allowed_authentication_flows,
			access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			access_token_type, access_token_algorithm, access_token_mapping,
			id_token_algorithm, id_token_mapping, redirect_uris, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
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

	row := p.db.QueryRow(ctx, `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, created_at, updated_at
		FROM applications 
		WHERE tenant = $1 AND realm = $2 AND client_id = $3
	`, tenant, realm, id)

	var app model.Application
	var allowedScopesJSON, allowedGrantsJSON, allowedAuthFlowsJSON []byte
	var accessTokenMappingJSON, idTokenMappingJSON, redirectUrisJSON []byte

	err := row.Scan(
		&app.Tenant,
		&app.Realm,
		&app.ClientId,
		&app.ClientSecret,
		&app.Confidential,
		&app.ConsentRequired,
		&app.Description,
		&allowedScopesJSON,
		&allowedGrantsJSON,
		&allowedAuthFlowsJSON,
		&app.AccessTokenLifetime,
		&app.RefreshTokenLifetime,
		&app.IdTokenLifetime,
		&app.AccessTokenType,
		&app.AccessTokenAlgorithm,
		&accessTokenMappingJSON,
		&app.IdTokenAlgorithm,
		&idTokenMappingJSON,
		&redirectUrisJSON,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(allowedScopesJSON, &app.AllowedScopes)
	json.Unmarshal(allowedGrantsJSON, &app.AllowedGrants)
	json.Unmarshal(allowedAuthFlowsJSON, &app.AllowedAuthenticationFlows)
	json.Unmarshal(accessTokenMappingJSON, &app.AccessTokenMapping)
	json.Unmarshal(idTokenMappingJSON, &app.IdTokenMapping)
	json.Unmarshal(redirectUrisJSON, &app.RedirectUris)

	return &app, nil
}

func (p *PostgresApplicationDB) UpdateApplication(ctx context.Context, app *model.Application) error {
	now := time.Now()

	// Convert JSON fields to strings
	allowedScopesJSON, _ := json.Marshal(app.AllowedScopes)
	allowedGrantsJSON, _ := json.Marshal(app.AllowedGrants)
	allowedAuthFlowsJSON, _ := json.Marshal(app.AllowedAuthenticationFlows)
	accessTokenMappingJSON, _ := json.Marshal(app.AccessTokenMapping)
	idTokenMappingJSON, _ := json.Marshal(app.IdTokenMapping)
	redirectUrisJSON, _ := json.Marshal(app.RedirectUris)

	_, err := p.db.Exec(ctx, `
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
			updated_at = $17
		WHERE tenant = $18 AND realm = $19 AND client_id = $20
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
		now,
		app.Tenant,
		app.Realm,
		app.ClientId,
	)
	return err
}

func (p *PostgresApplicationDB) ListApplications(ctx context.Context, tenant, realm string) ([]model.Application, error) {
	rows, err := p.db.Query(ctx, `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, created_at, updated_at
		FROM applications 
		WHERE tenant = $1 AND realm = $2
	`, tenant, realm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		var allowedScopesJSON, allowedGrantsJSON, allowedAuthFlowsJSON []byte
		var accessTokenMappingJSON, idTokenMappingJSON, redirectUrisJSON []byte

		err := rows.Scan(
			&app.Tenant,
			&app.Realm,
			&app.ClientId,
			&app.ClientSecret,
			&app.Confidential,
			&app.ConsentRequired,
			&app.Description,
			&allowedScopesJSON,
			&allowedGrantsJSON,
			&allowedAuthFlowsJSON,
			&app.AccessTokenLifetime,
			&app.RefreshTokenLifetime,
			&app.IdTokenLifetime,
			&app.AccessTokenType,
			&app.AccessTokenAlgorithm,
			&accessTokenMappingJSON,
			&app.IdTokenAlgorithm,
			&idTokenMappingJSON,
			&redirectUrisJSON,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		json.Unmarshal(allowedScopesJSON, &app.AllowedScopes)
		json.Unmarshal(allowedGrantsJSON, &app.AllowedGrants)
		json.Unmarshal(allowedAuthFlowsJSON, &app.AllowedAuthenticationFlows)
		json.Unmarshal(accessTokenMappingJSON, &app.AccessTokenMapping)
		json.Unmarshal(idTokenMappingJSON, &app.IdTokenMapping)
		json.Unmarshal(redirectUrisJSON, &app.RedirectUris)

		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

func (p *PostgresApplicationDB) ListAllApplications(ctx context.Context) ([]model.Application, error) {
	rows, err := p.db.Query(ctx, `
		SELECT tenant, realm, client_id, client_secret, confidential, consent_required,
		       description, allowed_scopes, allowed_grants, allowed_authentication_flows,
		       access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
		       access_token_type, access_token_algorithm, access_token_mapping,
		       id_token_algorithm, id_token_mapping, redirect_uris, created_at, updated_at
		FROM applications
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		var allowedScopesJSON, allowedGrantsJSON, allowedAuthFlowsJSON []byte
		var accessTokenMappingJSON, idTokenMappingJSON, redirectUrisJSON []byte

		err := rows.Scan(
			&app.Tenant,
			&app.Realm,
			&app.ClientId,
			&app.ClientSecret,
			&app.Confidential,
			&app.ConsentRequired,
			&app.Description,
			&allowedScopesJSON,
			&allowedGrantsJSON,
			&allowedAuthFlowsJSON,
			&app.AccessTokenLifetime,
			&app.RefreshTokenLifetime,
			&app.IdTokenLifetime,
			&app.AccessTokenType,
			&app.AccessTokenAlgorithm,
			&accessTokenMappingJSON,
			&app.IdTokenAlgorithm,
			&idTokenMappingJSON,
			&redirectUrisJSON,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		json.Unmarshal(allowedScopesJSON, &app.AllowedScopes)
		json.Unmarshal(allowedGrantsJSON, &app.AllowedGrants)
		json.Unmarshal(allowedAuthFlowsJSON, &app.AllowedAuthenticationFlows)
		json.Unmarshal(accessTokenMappingJSON, &app.AccessTokenMapping)
		json.Unmarshal(idTokenMappingJSON, &app.IdTokenMapping)
		json.Unmarshal(redirectUrisJSON, &app.RedirectUris)

		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

func (p *PostgresApplicationDB) DeleteApplication(ctx context.Context, tenant, realm, id string) error {
	_, err := p.db.Exec(ctx, `
		DELETE FROM applications 
		WHERE tenant = $1 AND realm = $2 AND client_id = $3
	`, tenant, realm, id)
	return err
}
