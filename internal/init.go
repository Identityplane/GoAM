package internal

import (
	"goiam/internal/realms"
	"log"
	"os"
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
}

func getConfigPath() string {
	path := os.Getenv("GOIAM_CONFIG_PATH")
	if path == "" {
		path = "../config" // fallback for local dev
	}
	log.Printf("Using config path: %s", path)
	return path
}
