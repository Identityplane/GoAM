package postgres_adapter

import (
	"context"
	"fmt"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"

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
		log := logger.GetLogger()
		log.Debug().Err(err).Msg("warning: failed to check if realms table exists")
	}

	return &PostgresRealmDB{db: db}, nil
}

func (p *PostgresRealmDB) CreateRealm(ctx context.Context, realm model.Realm) error {

	// If realm settings are null we init a empty map
	if realm.RealmSettings == nil {
		realm.RealmSettings = make(map[string]string)
	}

	query := `
		INSERT INTO realms (
			tenant, realm, realm_name, base_url, realm_settings
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := p.db.Exec(ctx, query,
		realm.Tenant, realm.Realm, realm.RealmName, realm.BaseUrl, realm.RealmSettings,
	)
	if err != nil {
		return fmt.Errorf("insert realm: %w", err)
	}

	return nil
}

func (p *PostgresRealmDB) GetRealm(ctx context.Context, tenant, realm string) (*model.Realm, error) {
	query := `
		SELECT * FROM realms
		WHERE tenant = $1 AND realm = $2
	`

	rows, err := p.db.Query(ctx, query, tenant, realm)
	if err != nil {
		return nil, fmt.Errorf("select realm: %w", err)
	}
	defer rows.Close()

	realmConfig, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.Realm])
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan realm: %w", err)
	}

	return realmConfig, nil
}

func (p *PostgresRealmDB) UpdateRealm(ctx context.Context, realm *model.Realm) error {

	// If realm settings are null we init a empty map
	if realm.RealmSettings == nil {
		realm.RealmSettings = make(map[string]string)
	}

	query := `
		UPDATE realms
		SET realm_name = $1, base_url = $2, realm_settings = $3
		WHERE tenant = $4 AND realm = $5
	`

	result, err := p.db.Exec(ctx, query,
		realm.RealmName, realm.BaseUrl, realm.RealmSettings, realm.Tenant, realm.Realm,
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
		SELECT * FROM realms
		WHERE tenant = $1
	`

	rows, err := p.db.Query(ctx, query, tenant)
	if err != nil {
		return nil, fmt.Errorf("select realms: %w", err)
	}
	defer rows.Close()

	realms, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Realm])
	if err != nil {
		return nil, fmt.Errorf("collect realms: %w", err)
	}

	return realms, nil
}

func (p *PostgresRealmDB) ListAllRealms(ctx context.Context) ([]model.Realm, error) {
	query := `
		SELECT * FROM realms
	`

	rows, err := p.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select realms: %w", err)
	}
	defer rows.Close()

	realms, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Realm])
	if err != nil {
		return nil, fmt.Errorf("collect realms: %w", err)
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
