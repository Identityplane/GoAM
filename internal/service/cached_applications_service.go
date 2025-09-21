package service

import (
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

const (
	// applicationCacheTTL is the time-to-live for application cache entries
	applicationCacheTTL = 5 * time.Second
)

// cachedApplicationService implements ApplicationService with caching
type cachedApplicationService struct {
	applicationService services_interface.ApplicationService
	cache              services_interface.CacheService
}

// NewCachedApplicationService creates a new cached application service
func NewCachedApplicationService(applicationService services_interface.ApplicationService, cache services_interface.CacheService) services_interface.ApplicationService {
	return &cachedApplicationService{
		applicationService: applicationService,
		cache:              cache,
	}
}

// getApplicationCacheKey returns a cache key in the format /<tenant>/<realm>/application/<client_id>
func (s *cachedApplicationService) getApplicationCacheKey(tenant, realm, clientId string) string {
	return fmt.Sprintf("/%s/%s/application/%s", tenant, realm, clientId)
}

func (s *cachedApplicationService) GetApplication(tenant, realm, id string) (*model.Application, bool) {
	// Try to get from cache first
	cacheKey := s.getApplicationCacheKey(tenant, realm, id)
	if cached, found := s.cache.Get(cacheKey); found && cached != nil {
		if app, ok := cached.(*model.Application); ok {
			return app, true
		}
	}

	// If not in cache, get from service
	app, found := s.applicationService.GetApplication(tenant, realm, id)
	if found {
		// Cache the application
		if err := s.cache.Cache(cacheKey, app, applicationCacheTTL, 1); err != nil {
			log := logger.GetGoamLogger()
			log.Info().Err(err).Msg("failed to cache application")
		}
	}
	return app, found
}

// Direct pass-through methods (no caching)
func (s *cachedApplicationService) ListApplications(tenant, realm string) ([]model.Application, error) {
	return s.applicationService.ListApplications(tenant, realm)
}

// Direct pass-through methods (no caching)
func (s *cachedApplicationService) ListAllApplications() ([]model.Application, error) {
	return s.applicationService.ListAllApplications()
}

func (s *cachedApplicationService) CreateApplication(tenant, realm string, application model.Application) error {
	err := s.applicationService.CreateApplication(tenant, realm, application)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(tenant, realm, application.ClientId)
	return nil
}

func (s *cachedApplicationService) UpdateApplication(tenant, realm string, application model.Application) error {
	err := s.applicationService.UpdateApplication(tenant, realm, application)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(tenant, realm, application.ClientId)
	return nil
}

func (s *cachedApplicationService) DeleteApplication(tenant, realm, clientId string) error {
	err := s.applicationService.DeleteApplication(tenant, realm, clientId)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(tenant, realm, clientId)
	return nil
}

func (s *cachedApplicationService) RegenerateClientSecret(tenant, realm, clientId string) (string, error) {
	secret, err := s.applicationService.RegenerateClientSecret(tenant, realm, clientId)
	if err != nil {
		return "", err
	}

	// Invalidate cache since client secret has changed
	s.invalidateCache(tenant, realm, clientId)
	return secret, nil
}

func (s *cachedApplicationService) VerifyClientSecret(tenant, realm, clientId, clientSecret string) (bool, error) {
	return s.applicationService.VerifyClientSecret(tenant, realm, clientId, clientSecret)
}

// invalidateCache invalidates the application cache entry
func (s *cachedApplicationService) invalidateCache(tenant, realm, clientId string) {
	cacheKey := s.getApplicationCacheKey(tenant, realm, clientId)
	s.cache.Invalidate(cacheKey)
}
