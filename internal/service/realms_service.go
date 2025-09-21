package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/db"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"

	"github.com/pkg/errors"
)

// Intermediate used for deserialization
type flowRealmYaml struct {
	Name string `yaml:"name"`
}

func NewLoadedRealm(realmConfig *model.Realm, repos model.Repositories) *services_interface.LoadedRealm {

	realmId := realmConfig.Tenant + "/" + realmConfig.Realm

	return &services_interface.LoadedRealm{
		Config:       realmConfig,
		RealmID:      realmId,
		Repositories: &repos,
	}
}

// realmServiceImpl implements RealmService
type realmServiceImpl struct {
	realmDb         db.RealmDB
	userDb          db.UserDB
	userAttributeDb db.UserAttributeDB
}

// NewRealmService creates a new RealmService instance
func NewRealmService(realmDb db.RealmDB, userDb db.UserDB, userAttributeDb db.UserAttributeDB) services_interface.RealmService {
	return &realmServiceImpl{
		realmDb:         realmDb,
		userDb:          userDb,
		userAttributeDb: userAttributeDb,
	}
}

func (s *realmServiceImpl) GetRealm(tenant, realm string) (*services_interface.LoadedRealm, bool) {
	log := logger.GetGoamLogger()

	// Use the database to get the realm config
	realmConfig, err := s.realmDb.GetRealm(context.Background(), tenant, realm)

	if err != nil {
		log.Debug().Str("tenant", tenant).Str("realm", realm).Msg("cannot load realm")
		return nil, false
	}

	if realmConfig == nil {
		log.Debug().Msg("realm not found")
		return nil, false
	}

	// Load the realm with repo
	userRepo := repository.NewUserRepository(tenant, realm, s.userDb, s.userAttributeDb)
	emailSender := repository.NewDefaultEmailSender()
	repos := &model.Repositories{
		UserRepo:    userRepo,
		EmailSender: emailSender,
	}
	loadedRealm := NewLoadedRealm(realmConfig, *repos)

	return loadedRealm, true
}

func (s *realmServiceImpl) GetAllRealms() (map[string]*services_interface.LoadedRealm, error) {

	realmConfigs, err := s.realmDb.ListAllRealms(context.Background())

	if err != nil {
		return nil, errors.Errorf("cannot load all realms")
	}

	loadedRealms := make(map[string]*services_interface.LoadedRealm)

	for _, realmConfig := range realmConfigs {

		// Load the realm with repo
		userRepo := repository.NewUserRepository(realmConfig.Tenant, realmConfig.Realm, s.userDb, s.userAttributeDb)
		emailSender := repository.NewDefaultEmailSender()
		repos := &model.Repositories{
			UserRepo:    userRepo,
			EmailSender: emailSender,
		}
		loadedRealm := NewLoadedRealm(&realmConfig, *repos)

		loadedRealms[loadedRealm.RealmID] = loadedRealm
	}

	return loadedRealms, nil
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

	// Update all fields that are provided
	if realm.RealmName != "" {
		existing.RealmName = realm.RealmName
	}
	if realm.BaseUrl != "" {
		existing.BaseUrl = realm.BaseUrl
	}
	if realm.RealmSettings != nil {
		existing.RealmSettings = realm.RealmSettings
	}

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

func (s *realmServiceImpl) IsTenantNameAvailable(tenantName string) (bool, error) {

	// TODO we should optimize this by implementing this in the database
	existing, err := s.realmDb.ListRealms(context.Background(), tenantName)
	if err != nil {
		return false, fmt.Errorf("check existing realm: %w", err)
	}

	isAvailable := len(existing) == 0

	return isAvailable, nil
}
