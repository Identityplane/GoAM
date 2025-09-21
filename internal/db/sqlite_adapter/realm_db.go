package sqlite_adapter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/jmoiron/sqlx"
)

type SQLiteRealmDB struct {
	db *sqlx.DB
}

func NewRealmDB(db *sql.DB) (*SQLiteRealmDB, error) {
	// Wrap with sqlx
	sqlxDB := sqlx.NewDb(db, "sqlite3")

	// Check if the connection works and realms table exists by executing a query
	_, err := sqlxDB.Exec(`
		SELECT 1 FROM realms LIMIT 1
	`)
	if err != nil {
		log := logger.GetGoamLogger()
		log.Debug().Err(err).Msg("warning: failed to check if realms table exists")
	}

	return &SQLiteRealmDB{db: sqlxDB}, nil
}

func (s *SQLiteRealmDB) CreateRealm(ctx context.Context, realm model.Realm) error {
	// Convert RealmSettings map to JSON string
	realmSettingsJSON, _ := json.Marshal(realm.RealmSettings)

	query := `
		INSERT INTO realms (
			tenant, realm, realm_name, base_url, realm_settings
		) VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		realm.Tenant, realm.Realm, realm.RealmName, realm.BaseUrl, string(realmSettingsJSON),
	)
	if err != nil {
		return fmt.Errorf("insert realm: %w", err)
	}

	return nil
}

func (s *SQLiteRealmDB) GetRealm(ctx context.Context, tenant, realm string) (*model.Realm, error) {
	query := `
		SELECT tenant, realm, realm_name, base_url, realm_settings
		FROM realms
		WHERE tenant = ? AND realm = ?
	`

	var realmConfig model.Realm
	var realmSettingsJSON string
	err := s.db.QueryRowContext(ctx, query, tenant, realm).Scan(
		&realmConfig.Tenant, &realmConfig.Realm,
		&realmConfig.RealmName, &realmConfig.BaseUrl, &realmSettingsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select realm: %w", err)
	}

	// Parse RealmSettings JSON
	realmConfig.RealmSettings = make(map[string]string)
	if realmSettingsJSON != "" && realmSettingsJSON != "null" {
		_ = json.Unmarshal([]byte(realmSettingsJSON), &realmConfig.RealmSettings)
	}

	return &realmConfig, nil
}

func (s *SQLiteRealmDB) UpdateRealm(ctx context.Context, realm *model.Realm) error {
	// Convert RealmSettings map to JSON string
	realmSettingsJSON, _ := json.Marshal(realm.RealmSettings)

	query := `
		UPDATE realms
		SET realm_name = ?, base_url = ?, realm_settings = ?
		WHERE tenant = ? AND realm = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		realm.RealmName, realm.BaseUrl, string(realmSettingsJSON), realm.Tenant, realm.Realm,
	)
	if err != nil {
		return fmt.Errorf("update realm: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("realm not found")
	}

	return nil
}

func (s *SQLiteRealmDB) ListRealms(ctx context.Context, tenant string) ([]model.Realm, error) {
	query := `
		SELECT tenant, realm, realm_name, base_url, realm_settings
		FROM realms
		WHERE tenant = ?
	`

	rows, err := s.db.QueryContext(ctx, query, tenant)
	if err != nil {
		return nil, fmt.Errorf("select realms: %w", err)
	}
	defer rows.Close()

	var realms []model.Realm
	for rows.Next() {
		var realm model.Realm
		var realmSettingsJSON string
		err := rows.Scan(
			&realm.Tenant, &realm.Realm,
			&realm.RealmName, &realm.BaseUrl, &realmSettingsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan realm: %w", err)
		}

		// Parse RealmSettings JSON
		realm.RealmSettings = make(map[string]string)
		if realmSettingsJSON != "" && realmSettingsJSON != "null" {
			_ = json.Unmarshal([]byte(realmSettingsJSON), &realm.RealmSettings)
		}

		realms = append(realms, realm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate realms: %w", err)
	}

	return realms, nil
}

func (s *SQLiteRealmDB) ListAllRealms(ctx context.Context) ([]model.Realm, error) {
	query := `
		SELECT tenant, realm, realm_name, base_url, realm_settings
		FROM realms
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select realms: %w", err)
	}
	defer rows.Close()

	var realms []model.Realm
	for rows.Next() {
		var realm model.Realm
		var realmSettingsJSON string
		err := rows.Scan(
			&realm.Tenant, &realm.Realm,
			&realm.RealmName, &realm.BaseUrl, &realmSettingsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan realm: %w", err)
		}

		// Parse RealmSettings JSON
		realm.RealmSettings = make(map[string]string)
		if realmSettingsJSON != "" && realmSettingsJSON != "null" {
			_ = json.Unmarshal([]byte(realmSettingsJSON), &realm.RealmSettings)
		}

		realms = append(realms, realm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate realms: %w", err)
	}

	return realms, nil
}

func (s *SQLiteRealmDB) DeleteRealm(ctx context.Context, tenant, realm string) error {
	query := `
		DELETE FROM realms
		WHERE tenant = ? AND realm = ?
	`

	result, err := s.db.ExecContext(ctx, query, tenant, realm)
	if err != nil {
		return fmt.Errorf("delete realm: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("realm not found")
	}

	return nil
}
