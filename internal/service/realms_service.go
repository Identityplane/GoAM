package service

import (
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/db"
	"goiam/internal/logger"
	"goiam/internal/model"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// RealmService defines the business logic for realm operations
type RealmService interface {
	// GetRealm returns a loaded realm configuration by its composite ID
	GetRealm(id string) (*LoadedRealm, bool)
	// InitRealms loads all static realm configurations from disk
	InitRealms(configRoot string, userDb db.UserDB) error
	// GetAllRealms returns all loaded realms
	GetAllRealms() map[string]*LoadedRealm
}

// Intermediate used for deserialization
type flowRealmYaml struct {
	Name string `yaml:"name"`
}

// LoadedRealm wraps a RealmConfig with metadata for tracking its source.
type LoadedRealm struct {
	Config       *model.RealmConfig       // parsed realm config
	RealmID      string                   // composite ID like "acme/customers"
	Path         string                   // original file path, useful for debugging/reloads
	Repositories *repository.Repositories // services for this realm
}

// realmServiceImpl implements RealmService
type realmServiceImpl struct {
	loadedRealms   map[string]*LoadedRealm
	loadedRealmsMu sync.RWMutex
}

// NewRealmService creates a new RealmService instance
func NewRealmService() RealmService {
	return &realmServiceImpl{
		loadedRealms: make(map[string]*LoadedRealm),
	}
}

func (s *realmServiceImpl) InitRealms(configRoot string, userDb db.UserDB) error {

	// Swap global registry
	s.loadedRealmsMu.Lock()
	defer s.loadedRealmsMu.Unlock()

	// Load realms from config directory
	newRealms, err := loadRealmsFromConfigDir(configRoot)
	if err != nil {
		return fmt.Errorf("failed to load realms from config directory %s: %w", configRoot, err)
	}

	// Init services for each realm
	for _, realm := range newRealms {

		// Init services for realm
		realm.Repositories = &repository.Repositories{}

		// Init user repository
		realm.Repositories.UserRepo = db.NewUserRepository(realm.Config.Tenant, realm.Config.Realm, userDb)
		logger.DebugNoContext("Initialized user repository for realm %s", realm.RealmID)
	}

	s.loadedRealms = newRealms

	return nil
}

// loadRealmsFromConfigDir loads all realm configurations from the given config root directory
func loadRealmsFromConfigDir(configRoot string) (map[string]*LoadedRealm, error) {

	newRealms := make(map[string]*LoadedRealm)

	tenantsPath := filepath.Join(configRoot, "tenants")
	logger.DebugNoContext("Walking config dir: %s", tenantsPath)

	// We need this to calculate the depth of the current path
	baseDepth := strings.Count(tenantsPath, string(os.PathSeparator))

	// Walk the config directory
	err := filepath.WalkDir(tenantsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil // Ignore non-yaml files
		}

		// Skip if the depth is greater than 2
		currentDepth := strings.Count(path, string(os.PathSeparator)) - baseDepth
		if currentDepth > 2 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		logger.DebugNoContext("Loading realm config: %s\n", path)

		cfg, err := loadRealmConfigFromFilePath(path)
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
		return nil, fmt.Errorf("failed to walk realm config directory %s: %w", configRoot, err)
	}

	return newRealms, nil

}

func (s *realmServiceImpl) GetRealm(id string) (*LoadedRealm, bool) {
	s.loadedRealmsMu.RLock()
	defer s.loadedRealmsMu.RUnlock()
	r, ok := s.loadedRealms[id]
	return r, ok
}

func (s *realmServiceImpl) GetAllRealms() map[string]*LoadedRealm {
	s.loadedRealmsMu.RLock()
	defer s.loadedRealmsMu.RUnlock()

	// Shallow copy to prevent external mutation
	copy := make(map[string]*LoadedRealm, len(s.loadedRealms))
	for k, v := range s.loadedRealms {
		copy[k] = v
	}
	return copy
}

// Helper function to load realm config from file
// Does not load flows, only realm config as we have a seperate service for loading flows
func loadRealmConfigFromFilePath(path string) (*model.RealmConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var unmarshaledFlowYaml struct {
		Realm string `yaml:"realm"`
	}

	if err := yaml.Unmarshal(data, &unmarshaledFlowYaml); err != nil {
		return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
	}

	if unmarshaledFlowYaml.Realm == "" {
		return nil, fmt.Errorf("invalid config in %s: 'realm' is required", path)
	}

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

	return &model.RealmConfig{
		Realm:  unmarshaledFlowYaml.Realm,
		Tenant: tenant,
	}, nil
}
