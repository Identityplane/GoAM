package internal

import (
	"goiam/internal/realms"
	"log"
)

var (
	// All loaded realm configurations, indexed by "tenant/realm"
	LoadedRealms = map[string]*realms.LoadedRealm{}
)

// Initialize loads all tenant/realm configurations at startup.
// Each realm must include its own flow configuration.
func Initialize() {
	const realmConfigRoot = "../config/tenants"

	if err := realms.InitRealms(realmConfigRoot); err != nil {
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
