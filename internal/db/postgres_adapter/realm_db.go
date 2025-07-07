package postgres_adapter

import (
	"context"
	"fmt"

	"github.com/gianlucafrei/GoAM/internal/logger"
	"github.com/gianlucafrei/GoAM/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRealmDB struct {
	db *pgxpool.Pool
}

func NewPostgresRealmDB(db *pgxpool.Pool) (*PostgresRealmDB, error) {
	// Check if the connection works and realms table exists by executing a query
	_, err := db.Exec(context.Background(), `
		SELECT 1 FROM realms LIMIT 1
	`)
	if err != nil {
		logger.DebugNoContext("Warning: failed to check if realms table exists: %v", err)
	}

	return &PostgresRealmDB{db: db}, nil
}

func (p *PostgresRealmDB) CreateRealm(ctx context.Context, realm model.Realm) error {
	query := `
		INSERT INTO realms (
			tenant, realm, realm_name, base_url
		) VALUES ($1, $2, $3, $4)
	`

	_, err := p.db.Exec(ctx, query,
		realm.Tenant, realm.Realm, realm.RealmName, realm.BaseUrl,
	)
	if err != nil {
		return fmt.Errorf("insert realm: %w", err)
	}

	return nil
}

func (p *PostgresRealmDB) GetRealm(ctx context.Context, tenant, realm string) (*model.Realm, error) {
	query := `
		SELECT tenant, realm, realm_name, base_url
		FROM realms
		WHERE tenant = $1 AND realm = $2
	`

	var realmConfig model.Realm
	err := p.db.QueryRow(ctx, query, tenant, realm).Scan(
		&realmConfig.Tenant, &realmConfig.Realm,
		&realmConfig.RealmName, &realmConfig.BaseUrl,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select realm: %w", err)
	}

	return &realmConfig, nil
}

func (p *PostgresRealmDB) UpdateRealm(ctx context.Context, realm *model.Realm) error {
	query := `
		UPDATE realms
		SET realm_name = $1, base_url = $2
		WHERE tenant = $3 AND realm = $4
	`

	result, err := p.db.Exec(ctx, query,
		realm.RealmName, realm.BaseUrl, realm.Tenant, realm.Realm,
	)
	if err != nil {
		return fmt.Errorf("update realm: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("realm not found")
	}

	return nil
}

func (p *PostgresRealmDB) ListRealms(ctx context.Context, tenant string) ([]model.Realm, error) {
	query := `
		SELECT tenant, realm, realm_name, base_url
		FROM realms
		WHERE tenant = $1
	`

	rows, err := p.db.Query(ctx, query, tenant)
	if err != nil {
		return nil, fmt.Errorf("select realms: %w", err)
	}
	defer rows.Close()

	var realms []model.Realm
	for rows.Next() {
		var realm model.Realm
		err := rows.Scan(
			&realm.Tenant, &realm.Realm,
			&realm.RealmName, &realm.BaseUrl,
		)
		if err != nil {
			return nil, fmt.Errorf("scan realm: %w", err)
		}

		realms = append(realms, realm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate realms: %w", err)
	}

	return realms, nil
}

func (p *PostgresRealmDB) ListAllRealms(ctx context.Context) ([]model.Realm, error) {
	query := `
		SELECT tenant, realm, realm_name, base_url
		FROM realms
	`

	rows, err := p.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select realms: %w", err)
	}
	defer rows.Close()

	var realms []model.Realm
	for rows.Next() {
		var realm model.Realm
		err := rows.Scan(
			&realm.Tenant, &realm.Realm,
			&realm.RealmName, &realm.BaseUrl,
		)
		if err != nil {
			return nil, fmt.Errorf("scan realm: %w", err)
		}

		realms = append(realms, realm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate realms: %w", err)
	}

	return realms, nil
}

func (p *PostgresRealmDB) DeleteRealm(ctx context.Context, tenant, realm string) error {
	query := `
		DELETE FROM realms
		WHERE tenant = $1 AND realm = $2
	`

	result, err := p.db.Exec(ctx, query, tenant, realm)
	if err != nil {
		return fmt.Errorf("delete realm: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("realm not found")
	}

	return nil
}
