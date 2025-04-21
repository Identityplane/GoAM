package internal

import (
	"database/sql"
	"goiam/internal/auth/graph"
	"goiam/internal/auth/repository"
	"goiam/internal/config"
	"goiam/internal/db/postgres_adapter"
	"goiam/internal/db/service"
	"goiam/internal/db/sqlite_adapter"
	"goiam/internal/logger"
	"goiam/internal/realms"
	"goiam/internal/web"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/valyala/fasthttp"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms     = map[string]*realms.LoadedRealm{}
	UserAdminService service.UserAdminService
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize() {

	// Prinout config path
	logger.DebugNoContext("Using config path: %s", config.ConfigPath)
	if err := realms.InitRealms(config.ConfigPath); err != nil {
		logger.PanicNoContext("failed to initialize realms: %v", err)
	}

	// Cache all loaded realms locally if needed
	allRealms := make(map[string]*realms.LoadedRealm)
	for id, realm := range realms.GetAllRealms() {
		allRealms[id] = realm
	}

	LoadedRealms = allRealms
	logger.DebugNoContext("Loaded %d realms\n", len(LoadedRealms))

	// Init Database
	// Currenlty we only support sqlite but later we need to load this from the config
	var db any
	var err error
	if strings.HasPrefix(config.DBConnString, "postgres://") {

		db, err = postgres_adapter.Init(postgres_adapter.Config{
			Driver: "postgres",
			DSN:    config.DBConnString,
		})

		logger.DebugNoContext("Initializing postgres database")
	} else {
		db, err = sqlite_adapter.Init(sqlite_adapter.Config{
			Driver: "sqlite",
			DSN:    config.DBConnString,
		})

		logger.DebugNoContext("Initializing sqlite database")
	}

	if err != nil {
		logger.PanicNoContext("DB init failed: %v", err)
		return
	}

	// Iterate over all tenants and realms and initialize a user repository
	// Currently each realm has a unique key, so we can just use that
	for _, realm := range LoadedRealms {

		var userRepo repository.UserRepository

		// case for sqllite and postgres
		switch dbTyped := db.(type) {
		case *sql.DB: // sqlite
			userDb, err := sqlite_adapter.NewSQLiteUserDB(dbTyped)

			if err != nil {
				logger.PanicNoContext("Failed to create sqlite user db: %v", err)
			}

			config.SqliteUserDB = userDb
			userRepo = sqlite_adapter.NewUserRepository(realm.Config.Tenant, realm.Config.Realm, userDb)

		case *pgx.Conn:
			userDb, err := postgres_adapter.NewPostgresUserDB(dbTyped)
			if err != nil {
				logger.PanicNoContext("Failed to create postgres user db: %v", err)
			}

			config.PostgresUserDB = userDb
			userRepo = postgres_adapter.NewUserRepository(realm.Config.Tenant, realm.Config.Realm, userDb)
		}

		logger.DebugNoContext("Initialized user repository for realm %s/%s", realm.Config.Tenant, realm.Config.Realm)

		// Init the service registry for this realm
		realm.Services = &graph.ServiceRegistry{
			UserRepo: userRepo,
		}
	}

	// Init user service for admin api
	dbDriverName := config.GetDbDriverName()
	switch dbDriverName {
	case "postgres":
		UserAdminService = service.NewUserService(config.PostgresUserDB)
	case "sqlite":
		UserAdminService = service.NewUserService(config.SqliteUserDB)
	default:
		logger.PanicNoContext("Unsupported database driver: %s", dbDriverName)
	}

	// Init web adapter
	r := web.New(UserAdminService)

	logger.DebugNoContext("Server running on http://localhost:8080")
	if err := fasthttp.ListenAndServe(":8080", web.TopLevelMiddleware(r.Handler)); err != nil {
		logger.PanicNoContext("Error: %s", err)
	}
}
