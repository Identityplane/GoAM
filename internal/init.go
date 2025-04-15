package internal

import (
	"goiam/internal/auth/graph"
	"goiam/internal/db/model"
	"goiam/internal/db/sqlite_adapter"
	"goiam/internal/realms"
	"goiam/internal/web"
	"log"
	"os"

	"github.com/valyala/fasthttp"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms = map[string]*realms.LoadedRealm{}
	ConfigPath   = getConfigPath()
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize() {

	// Prinout config path
	log.Printf("Using config path: %s", ConfigPath)
	if err := realms.InitRealms(ConfigPath); err != nil {
		log.Fatalf("failed to initialize realms: %v", err)
	}

	// Cache all loaded realms locally if needed
	allRealms := make(map[string]*realms.LoadedRealm)
	for id, realm := range realms.GetAllRealms() {
		allRealms[id] = realm
	}

	LoadedRealms = allRealms
	log.Printf("Loaded %d realms\n", len(LoadedRealms))

	// Init Database
	// Currenlty we only support sqlite but later we need to load this from the config
	db, err := sqlite_adapter.Init(sqlite_adapter.Config{
		Driver: "sqlite",
		DSN:    "goiam.db?_foreign_keys=on",
	})
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
		return
	}

	// Iterate over all tenants and realms and initialize a user repository
	// Currently each realm has a unique key, so we can just use that
	for _, realm := range LoadedRealms {

		// Init user repository
		var userDb model.UserDB = sqlite_adapter.NewSQLiteUserDB(db)
		userRepo := sqlite_adapter.NewUserRepository(realm.Config.Tenant, realm.Config.Realm, userDb)
		log.Printf("Initialized user repository for realm %s/%s", realm.Config.Tenant, realm.Config.Realm)

		// Init the service registry for this realm
		realm.Services = &graph.ServiceRegistry{
			UserRepo: userRepo,
		}

	}

	// Init web adapter
	r := web.New(ConfigPath)

	log.Println("Server running on http://localhost:8080")
	if err := fasthttp.ListenAndServe(":8080", r.Handler); err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func getConfigPath() string {
	path := os.Getenv("GOIAM_CONFIG_PATH")
	if path == "" {
		path = "../config" // fallback for local dev
	}
	log.Printf("Using config path: %s", path)
	return path
}
