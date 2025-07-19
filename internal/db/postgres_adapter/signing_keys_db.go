package postgres_adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSigningKeysDB struct {
	db *pgxpool.Pool
}

// NewPostgresSigningKeysDB creates a new PostgresSigningKeysDB instance
func NewPostgresSigningKeysDB(db *pgxpool.Pool) (*PostgresSigningKeysDB, error) {
	// Check if the connection works and signing_keys table exists
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM signing_keys LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check if signing_keys table exists: %w", err)
	}

	return &PostgresSigningKeysDB{db: db}, nil
}

func (s *PostgresSigningKeysDB) CreateSigningKey(ctx context.Context, key model.SigningKey) error {
	query := `
		INSERT INTO signing_keys (
			tenant, realm, kid, active, algorithm, implementation,
			signing_key_material, public_key_jwk, created, disabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var disabledStr interface{}
	if key.Disabled != nil {
		disabledStr = key.Disabled.Format(time.RFC3339)
	} else {
		disabledStr = nil
	}

	_, err := s.db.Exec(ctx, query,
		key.Tenant,
		key.Realm,
		key.Kid,
		key.Active,
		key.Algorithm,
		key.Implementation,
		key.SigningKeyMaterial,
		key.PublicKeyJWK,
		key.Created.Format(time.RFC3339),
		disabledStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create signing key: %w", err)
	}

	return nil
}

func (s *PostgresSigningKeysDB) GetSigningKey(ctx context.Context, tenant, realm, kid string) (*model.SigningKey, error) {
	query := `
		SELECT tenant, realm, kid, active, algorithm, implementation,
		       signing_key_material, public_key_jwk, created, disabled
		FROM signing_keys
		WHERE tenant = $1 AND realm = $2 AND kid = $3
	`

	var key model.SigningKey
	var created time.Time
	var disabled *time.Time

	err := s.db.QueryRow(ctx, query, tenant, realm, kid).Scan(
		&key.Tenant,
		&key.Realm,
		&key.Kid,
		&key.Active,
		&key.Algorithm,
		&key.Implementation,
		&key.SigningKeyMaterial,
		&key.PublicKeyJWK,
		&created,
		&disabled,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	key.Created = created
	key.Disabled = disabled

	return &key, nil
}

func (s *PostgresSigningKeysDB) ListSigningKeys(ctx context.Context, tenant, realm string) ([]model.SigningKey, error) {
	query := `
		SELECT tenant, realm, kid, active, algorithm, implementation,
		       signing_key_material, public_key_jwk, created, disabled
		FROM signing_keys
		WHERE tenant = $1 AND realm = $2
	`

	rows, err := s.db.Query(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list signing keys: %w", err)
	}
	defer rows.Close()

	var keys []model.SigningKey
	for rows.Next() {
		var key model.SigningKey
		var created time.Time
		var disabled *time.Time

		err := rows.Scan(
			&key.Tenant,
			&key.Realm,
			&key.Kid,
			&key.Active,
			&key.Algorithm,
			&key.Implementation,
			&key.SigningKeyMaterial,
			&key.PublicKeyJWK,
			&created,
			&disabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan signing key: %w", err)
		}

		key.Created = created
		key.Disabled = disabled

		keys = append(keys, key)
	}

	return keys, nil
}

func (s *PostgresSigningKeysDB) ListActiveSigningKeys(ctx context.Context, tenant, realm string) ([]model.SigningKey, error) {
	query := `
		SELECT tenant, realm, kid, active, algorithm, implementation,
		       signing_key_material, public_key_jwk, created, disabled
		FROM signing_keys
		WHERE tenant = $1 AND realm = $2 AND active = true
	`

	rows, err := s.db.Query(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("failed to list active signing keys: %w", err)
	}
	defer rows.Close()

	var keys []model.SigningKey
	for rows.Next() {
		var key model.SigningKey
		var created time.Time
		var disabled *time.Time

		err := rows.Scan(
			&key.Tenant,
			&key.Realm,
			&key.Kid,
			&key.Active,
			&key.Algorithm,
			&key.Implementation,
			&key.SigningKeyMaterial,
			&key.PublicKeyJWK,
			&created,
			&disabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan signing key: %w", err)
		}

		key.Created = created
		key.Disabled = disabled

		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating signing keys: %w", err)
	}

	return keys, nil
}

func (s *PostgresSigningKeysDB) UpdateSigningKey(ctx context.Context, key *model.SigningKey) error {
	// First get the current key to preserve active status if not changed
	currentKey, err := s.GetSigningKey(ctx, key.Tenant, key.Realm, key.Kid)
	if err != nil {
		return fmt.Errorf("failed to get current signing key: %w", err)
	}
	if currentKey == nil {
		return fmt.Errorf("signing key not found")
	}

	// If active status is not explicitly set, preserve the current one
	if key.Active == false && currentKey.Active == true {
		key.Active = currentKey.Active
	}

	query := `
		UPDATE signing_keys SET
			active = $1,
			algorithm = $2,
			implementation = $3,
			signing_key_material = $4,
			public_key_jwk = $5,
			created = $6,
			disabled = $7
		WHERE tenant = $8 AND realm = $9 AND kid = $10
	`

	var disabledStr interface{}
	if key.Disabled != nil {
		disabledStr = key.Disabled.Format(time.RFC3339)
	} else {
		disabledStr = nil
	}

	result, err := s.db.Exec(ctx, query,
		key.Active,
		key.Algorithm,
		key.Implementation,
		key.SigningKeyMaterial,
		key.PublicKeyJWK,
		key.Created.Format(time.RFC3339),
		disabledStr,
		key.Tenant,
		key.Realm,
		key.Kid,
	)
	if err != nil {
		return fmt.Errorf("failed to update signing key: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no signing key found to update")
	}

	return nil
}

func (s *PostgresSigningKeysDB) DeleteSigningKey(ctx context.Context, tenant, realm, kid string) error {
	query := `
		DELETE FROM signing_keys
		WHERE tenant = $1 AND realm = $2 AND kid = $3
	`

	_, err := s.db.Exec(ctx, query, tenant, realm, kid)
	if err != nil {
		return fmt.Errorf("failed to delete signing key: %w", err)
	}

	return nil
}

func (s *PostgresSigningKeysDB) DisableSigningKey(ctx context.Context, tenant, realm, kid string) error {
	query := `
		UPDATE signing_keys
		SET active = false, disabled = $1
		WHERE tenant = $2 AND realm = $3 AND kid = $4
	`

	now := time.Now()
	_, err := s.db.Exec(ctx, query, now.Format(time.RFC3339), tenant, realm, kid)
	if err != nil {
		return fmt.Errorf("failed to disable signing key: %w", err)
	}

	return nil
}
