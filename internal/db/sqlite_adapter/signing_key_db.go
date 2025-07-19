package sqlite_adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/model"
)

type SQLiteSigningKeyDB struct {
	db *sql.DB
}

func NewSigningKeyDB(db *sql.DB) (*SQLiteSigningKeyDB, error) {
	// Check if the connection works and signing_keys table exists by executing a query
	_, err := db.Exec(`
		SELECT 1 FROM signing_keys LIMIT 1
	`)
	if err != nil {
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("warning: failed to check if signing_keys table exists")
	}

	return &SQLiteSigningKeyDB{db: db}, nil
}

func (s *SQLiteSigningKeyDB) CreateSigningKey(ctx context.Context, key model.SigningKey) error {
	query := `
		INSERT INTO signing_keys (
			tenant, realm, kid, active, algorithm, implementation,
			signing_key_material, public_key_jwk, created
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		key.Tenant, key.Realm, key.Kid, key.Active, key.Algorithm,
		key.Implementation, key.SigningKeyMaterial, key.PublicKeyJWK,
		key.Created,
	)
	if err != nil {
		return fmt.Errorf("insert signing key: %w", err)
	}

	return nil
}

func (s *SQLiteSigningKeyDB) GetSigningKey(ctx context.Context, tenant, realm, kid string) (*model.SigningKey, error) {
	query := `
		SELECT tenant, realm, kid, active, algorithm, implementation,
			signing_key_material, public_key_jwk, created, disabled
		FROM signing_keys
		WHERE tenant = ? AND realm = ? AND kid = ?
	`

	var key model.SigningKey
	var disabled sql.NullTime
	err := s.db.QueryRowContext(ctx, query, tenant, realm, kid).Scan(
		&key.Tenant, &key.Realm, &key.Kid, &key.Active, &key.Algorithm,
		&key.Implementation, &key.SigningKeyMaterial, &key.PublicKeyJWK,
		&key.Created, &disabled,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select signing key: %w", err)
	}

	if disabled.Valid {
		key.Disabled = &disabled.Time
	}

	return &key, nil
}

func (s *SQLiteSigningKeyDB) UpdateSigningKey(ctx context.Context, key *model.SigningKey) error {
	query := `
		UPDATE signing_keys
		SET active = ?, algorithm = ?, implementation = ?,
			signing_key_material = ?, public_key_jwk = ?
		WHERE tenant = ? AND realm = ? AND kid = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		key.Active, key.Algorithm, key.Implementation,
		key.SigningKeyMaterial, key.PublicKeyJWK,
		key.Tenant, key.Realm, key.Kid,
	)
	if err != nil {
		return fmt.Errorf("update signing key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("signing key not found")
	}

	return nil
}

func (s *SQLiteSigningKeyDB) ListSigningKeys(ctx context.Context, tenant, realm string) ([]model.SigningKey, error) {
	query := `
		SELECT tenant, realm, kid, active, algorithm, implementation,
			signing_key_material, public_key_jwk, created, disabled
		FROM signing_keys
		WHERE tenant = ? AND realm = ?
	`

	rows, err := s.db.QueryContext(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("select signing keys: %w", err)
	}
	defer rows.Close()

	var keys []model.SigningKey
	for rows.Next() {
		var key model.SigningKey
		var disabled sql.NullTime
		err := rows.Scan(
			&key.Tenant, &key.Realm, &key.Kid, &key.Active, &key.Algorithm,
			&key.Implementation, &key.SigningKeyMaterial, &key.PublicKeyJWK,
			&key.Created, &disabled,
		)
		if err != nil {
			return nil, fmt.Errorf("scan signing key: %w", err)
		}

		if disabled.Valid {
			key.Disabled = &disabled.Time
		}

		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signing keys: %w", err)
	}

	return keys, nil
}

func (s *SQLiteSigningKeyDB) ListActiveSigningKeys(ctx context.Context, tenant, realm string) ([]model.SigningKey, error) {
	query := `
		SELECT tenant, realm, kid, active, algorithm, implementation,
			signing_key_material, public_key_jwk, created, disabled
		FROM signing_keys
		WHERE tenant = ? AND realm = ? AND active = true
	`

	rows, err := s.db.QueryContext(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("select active signing keys: %w", err)
	}
	defer rows.Close()

	var keys []model.SigningKey
	for rows.Next() {
		var key model.SigningKey
		var disabled sql.NullTime
		err := rows.Scan(
			&key.Tenant, &key.Realm, &key.Kid, &key.Active, &key.Algorithm,
			&key.Implementation, &key.SigningKeyMaterial, &key.PublicKeyJWK,
			&key.Created, &disabled,
		)
		if err != nil {
			return nil, fmt.Errorf("scan signing key: %w", err)
		}

		if disabled.Valid {
			key.Disabled = &disabled.Time
		}

		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signing keys: %w", err)
	}

	return keys, nil
}

func (s *SQLiteSigningKeyDB) DisableSigningKey(ctx context.Context, tenant, realm, kid string) error {
	query := `
		UPDATE signing_keys
		SET active = false, disabled = ?
		WHERE tenant = ? AND realm = ? AND kid = ?
	`

	now := time.Now()
	result, err := s.db.ExecContext(ctx, query, now, tenant, realm, kid)
	if err != nil {
		return fmt.Errorf("disable signing key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("signing key not found")
	}

	return nil
}

func (s *SQLiteSigningKeyDB) DeleteSigningKey(ctx context.Context, tenant, realm, kid string) error {
	query := `
		DELETE FROM signing_keys
		WHERE tenant = ? AND realm = ? AND kid = ?
	`

	result, err := s.db.ExecContext(ctx, query, tenant, realm, kid)
	if err != nil {
		return fmt.Errorf("delete signing key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("signing key not found")
	}

	return nil
}
