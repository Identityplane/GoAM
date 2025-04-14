package realms

import (
	"errors"
	"fmt"
	"goiam/internal/auth/graph"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// RealmConfig represents the static configuration of a realm, typically loaded from YAML files.
// It includes the realm + tenant identifiers and a multiple FlowWithRoute (for now).
type RealmConfig struct {
	Realm  string          `yaml:"realm"`  // e.g. "customers"
	Tenant string          `yaml:"tenant"` // e.g. "acme"
	Flows  []FlowWithRoute `yaml:"flows"`  // now supports multiple flows
}

// Intermediate used for deserialization
type flowWithConfigPath struct {
	Route string `yaml:"route"`
	File  string `yaml:"file"`
}

// FlowWithRoute ties a graph flow definition to a public HTTP route.
type FlowWithRoute struct {
	Route string                // e.g. "/login"
	Flow  *graph.FlowDefinition // pre-loaded flow definition
}

// LoadedRealm wraps a RealmConfig with metadata for tracking its source.
type LoadedRealm struct {
	Config  *RealmConfig // parsed realm config
	RealmID string       // composite ID like "acme/customers"
	Path    string       // original file path, useful for debugging/reloads
}

var (
	// Internal registry of loaded realms (populated during InitRealms)
	loadedRealms   = make(map[string]*LoadedRealm)
	loadedRealmsMu sync.RWMutex
)

// InitRealms loads all static realm configurations from disk at startup.
// It recursively walks the provided configRoot (e.g. "config/tenants") and
// loads all files matching the pattern **/realms/*.yaml.
func InitRealms(configRoot string) error {
	newRealms := make(map[string]*LoadedRealm)

	err := filepath.WalkDir(configRoot, func(path string, d fs.DirEntry, err error) error {

		if err != nil || d.IsDir() || filepath.Ext(path) != ".yaml" {

			return nil // Ignore non-yaml files
		}

		fmt.Printf("Loading realm config: %s\n", path)

		cfg, err := loadRealmConfig(path)
		if err != nil {
			return fmt.Errorf("error loading realm config at %s: %w", path, err)
		}

		id := fmt.Sprintf("%s/%s", cfg.Tenant, cfg.Realm)
		newRealms[id] = &LoadedRealm{
			Config:  cfg,
			RealmID: id,
			Path:    path,
		}
		return nil
	})

	if err != nil {

		return fmt.Errorf("failed to walk realm config directory %s: %w", configRoot, err)
	}

	// Swap global registry
	loadedRealmsMu.Lock()
	defer loadedRealmsMu.Unlock()
	loadedRealms = newRealms

	return nil // All good
}
func loadRealmConfig(path string) (*RealmConfig, error) {
	data, err := os.ReadFile(path) // #nosec G304 (the path is trusted as it is not meant to be used with user input)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	// Parse raw YAML with flow paths
	var raw struct {
		Realm string               `yaml:"realm"`
		Flows []flowWithConfigPath `yaml:"flows"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
	}

	if raw.Realm == "" {
		return nil, fmt.Errorf("invalid config in %s: 'realm' is required", path)
	}
	if len(raw.Flows) == 0 {
		return nil, fmt.Errorf("invalid config in %s: at least one flow is required", path)
	}

	// Inject tenant from directory name: /tenants/{tenant}/realms/{realm}.yaml
	segments := strings.Split(filepath.ToSlash(path), "/")
	tenantIdx := -1
	for i, segment := range segments {
		if segment == "tenants" && i+1 < len(segments) {
			tenantIdx = i + 1
			break
		}
	}
	if tenantIdx == -1 {
		return nil, fmt.Errorf("could not infer tenant name from path: %s", path)
	}
	tenant := segments[tenantIdx]

	// Load flow files
	var loadedFlows []FlowWithRoute
	for _, entry := range raw.Flows {
		if entry.Route == "" || entry.File == "" {
			return nil, fmt.Errorf("invalid flow entry in %s: route and file required", path)
		}

		flowPath := filepath.Join(filepath.Dir(path), "..", "..", "flows", entry.File)
		flow, err := LoadFlowFromYAML(flowPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load flow %q: %w", flowPath, err)
		}

		loadedFlows = append(loadedFlows, FlowWithRoute{
			Route: entry.Route,
			Flow:  flow,
		})
	}

	return &RealmConfig{
		Realm:  raw.Realm,
		Tenant: tenant,
		Flows:  loadedFlows,
	}, nil
}

// GetRealm returns a loaded realm configuration by its composite ID (e.g. "acme/customers").
func GetRealm(id string) (*LoadedRealm, bool) {
	loadedRealmsMu.RLock()
	defer loadedRealmsMu.RUnlock()
	r, ok := loadedRealms[id]
	return r, ok
}
func LookupFlow(tenant, realm, path string) (*FlowWithRoute, error) {
	realmID := fmt.Sprintf("%s/%s", tenant, realm)
	loaded, ok := GetRealm(realmID)
	if !ok {
		return nil, errors.New("realm not found")
	}

	cleanPath := "/" + strings.TrimPrefix(path, "/")

	for _, f := range loaded.Config.Flows {
		if f.Route == cleanPath {
			return &f, nil
		}
	}

	return nil, errors.New("route not found in realm")
}

// ListFlowsPerRealm returns all flows defined for a given tenant + realm.
func ListFlowsPerRealm(tenant, realm string) ([]FlowWithRoute, error) {
	realmID := fmt.Sprintf("%s/%s", tenant, realm)
	loaded, ok := GetRealm(realmID)
	if !ok {
		return nil, fmt.Errorf("realm not found: %s", realmID)
	}
	return loaded.Config.Flows, nil
}

// LookupFlowByName finds a flow by its internal name (not route).
func LookupFlowByName(tenant, realm, name string) (*FlowWithRoute, error) {
	flows, err := ListFlowsPerRealm(tenant, realm)
	if err != nil {
		return nil, err
	}

	for _, f := range flows {
		if f.Flow != nil && f.Flow.Name == name {
			return &f, nil
		}
	}

	return nil, errors.New("flow not found: " + name)
}
