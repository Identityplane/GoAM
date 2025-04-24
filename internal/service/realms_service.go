package service

import (
	"context"
	"database/sql"
	"fmt"
	"goiam/internal/auth/repository"
	"goiam/internal/db"
	"goiam/internal/logger"
	"goiam/internal/model"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// RealmService defines the business logic for realm operations
type RealmService interface {
	// GetRealm returns a loaded realm configuration by its composite ID
	GetRealm(tenant, realm string) (*LoadedRealm, bool)
	// InitRealms loads all static realm configurations from disk
	InitRealms(configRoot string, userDb db.UserDB) error
	// GetAllRealms returns a map of all loaded realms with realmId as index
	GetAllRealms() (map[string]*LoadedRealm, error)
	// CreateRealm creates a new realm
	CreateRealm(realm *model.Realm) error
	// UpdateRealm updates an existing realm
	UpdateRealm(realm *model.Realm) error
	// DeleteRealm deletes a realm
	DeleteRealm(tenant, realm string) error
}

// Intermediate used for deserialization
type flowRealmYaml struct {
	Name string `yaml:"name"`
}

// LoadedRealm wraps a RealmConfig with metadata for tracking its source.
type LoadedRealm struct {
	Config       *model.Realm             // parsed realm config
	RealmID      string                   // composite ID like "acme/customers"
	Repositories *repository.Repositories // services for this realm
}

func NewLoadedRealm(realmConfig *model.Realm, userRepo repository.UserRepository) *LoadedRealm {

	realmId := realmConfig.Tenant + "/" + realmConfig.Realm
	repos := &repository.Repositories{UserRepo: userRepo}

	return &LoadedRealm{
		Config:       realmConfig,
		RealmID:      realmId,
		Repositories: repos,
	}
}

// realmServiceImpl implements RealmService
type realmServiceImpl struct {
	realmDb db.RealmDB
	userDb  db.UserDB
}

// NewRealmService creates a new RealmService instance
func NewRealmService(realmDb db.RealmDB, userDb db.UserDB) RealmService {
	return &realmServiceImpl{
		realmDb: realmDb,
		userDb:  userDb,
	}
}

func (s *realmServiceImpl) InitRealms(configRoot string, userDb db.UserDB) error {

	// Load realms from config directory
	realmsFromConfigDir, err := loadRealmsFromConfigDir(configRoot)
	if err != nil {
		return fmt.Errorf("failed to load realms from config directory %s: %w", configRoot, err)
	}

	// Init services for each realm
	for _, realm := range realmsFromConfigDir {

		// Init services for realm
		realm.Repositories = &repository.Repositories{}

		// Init user repository
		realm.Repositories.UserRepo = db.NewUserRepository(realm.Config.Tenant, realm.Config.Realm, userDb)
		logger.DebugNoContext("Initialized user repository for realm %s", realm.RealmID)
	}

	// Currently we store the local realms in the database
	for _, realm := range realmsFromConfigDir {

		s.realmDb.CreateRealm(context.Background(), *realm.Config)
	}

	return nil
}

func (s *realmServiceImpl) GetRealm(tenant, realm string) (*LoadedRealm, bool) {

	// Use the database to get the realm config
	realmConfig, err := s.realmDb.GetRealm(context.Background(), tenant, realm)

	if err != nil {
		logger.DebugNoContext("cannot load realm %s/%s", tenant, realm)
		return nil, false
	}

	if realmConfig == nil {
		return nil, false
	}

	// Load the realm with repo
	userRepo := db.NewUserRepository(tenant, realm, s.userDb)
	loadedRealm := NewLoadedRealm(realmConfig, userRepo)

	return loadedRealm, true
}

func (s *realmServiceImpl) GetAllRealms() (map[string]*LoadedRealm, error) {

	realmConfigs, err := s.realmDb.ListAllRealms(context.Background())

	if err != nil {
		return nil, errors.Errorf("cannot load all realms")
	}

	loadedRealms := make(map[string]*LoadedRealm)

	for _, realmConfig := range realmConfigs {

		// Load the realm with repo
		userRepo := db.NewUserRepository(realmConfig.Tenant, realmConfig.Tenant, s.userDb)
		loadedRealm := NewLoadedRealm(&realmConfig, userRepo)

		loadedRealms[loadedRealm.RealmID] = loadedRealm
	}

	return loadedRealms, nil
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
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk realm config directory %s: %w", configRoot, err)
	}

	return newRealms, nil

}

// Helper function to load realm config from file
// Does not load flows, only realm config as we have a seperate service for loading flows
func loadRealmConfigFromFilePath(path string) (*model.Realm, error) {
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

	return &model.Realm{
		Realm:  unmarshaledFlowYaml.Realm,
		Tenant: tenant,
	}, nil
}

func (s *realmServiceImpl) CreateRealm(realm *model.Realm) error {
	// Check if realm already exists
	existing, err := s.realmDb.GetRealm(context.Background(), realm.Tenant, realm.Realm)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check existing realm: %w", err)
	}
	if existing != nil {
		return nil // Idempotent - realm already exists
	}

	// Create new realm
	if err := s.realmDb.CreateRealm(context.Background(), *realm); err != nil {
		return fmt.Errorf("create realm: %w", err)
	}

	return nil
}

func (s *realmServiceImpl) UpdateRealm(realm *model.Realm) error {
	// Check if realm exists
	existing, err := s.realmDb.GetRealm(context.Background(), realm.Tenant, realm.Realm)
	if err != nil {
		return fmt.Errorf("check existing realm: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("realm not found")
	}

	// Only update realm_name
	existing.RealmName = realm.RealmName

	if err := s.realmDb.UpdateRealm(context.Background(), existing); err != nil {
		return fmt.Errorf("update realm: %w", err)
	}

	return nil
}

func (s *realmServiceImpl) DeleteRealm(tenant, realm string) error {
	// Check if realm exists
	existing, err := s.realmDb.GetRealm(context.Background(), tenant, realm)
	if err != nil {
		return fmt.Errorf("check existing realm: %w", err)
	}
	if existing == nil {
		return nil // Idempotent - realm already deleted
	}

	if err := s.realmDb.DeleteRealm(context.Background(), tenant, realm); err != nil {
		return fmt.Errorf("delete realm: %w", err)
	}

	return nil
}
