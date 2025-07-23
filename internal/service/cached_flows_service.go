package service

import (
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	// flowCacheTTL is the time-to-live for flow cache entries
	flowCacheTTL = 5 * time.Second
)

// cachedFlowService implements FlowService with caching
type cachedFlowService struct {
	flowService FlowService
	cache       CacheService
}

// NewCachedFlowService creates a new cached flow service
func NewCachedFlowService(flowService FlowService, cache CacheService) FlowService {
	return &cachedFlowService{
		flowService: flowService,
		cache:       cache,
	}
}

// getFlowByIdCacheKey returns a cache key in the format /<tenant>/<realm>/flow/id/<id>
func (s *cachedFlowService) getFlowByIdCacheKey(tenant, realm, id string) string {
	return fmt.Sprintf("/%s/%s/flow/id/%s", tenant, realm, id)
}

// getFlowByPathCacheKey returns a cache key in the format /<tenant>/<realm>/flow/path/<path>
func (s *cachedFlowService) getFlowByPathCacheKey(tenant, realm, path string) string {
	return fmt.Sprintf("/%s/%s/flow/path/%s", tenant, realm, path)
}

func (s *cachedFlowService) GetFlowById(tenant, realm, id string) (*model.Flow, bool) {
	log := logger.GetLogger()
	// Try to get from cache first
	cacheKey := s.getFlowByIdCacheKey(tenant, realm, id)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if flow, ok := cached.(*model.Flow); ok {
			return flow, true
		}
	}

	// If not in cache, get from service
	flow, exists := s.flowService.GetFlowById(tenant, realm, id)
	if !exists {
		return nil, false
	}

	// Cache the result
	err := s.cache.Cache(cacheKey, flow, flowCacheTTL, 1)
	if err != nil {
		log.Info().Err(err).Msg("failed to cache flow by id")
	}

	return flow, true
}

func (s *cachedFlowService) GetFlowByPath(tenant, realm, path string) (*model.Flow, bool) {
	log := logger.GetLogger()
	// Try to get from cache first
	cacheKey := s.getFlowByPathCacheKey(tenant, realm, path)
	if cached, exists := s.cache.Get(cacheKey); exists {
		if flow, ok := cached.(*model.Flow); ok {
			return flow, true
		}
	}

	// If not in cache, get from service
	flow, exists := s.flowService.GetFlowByPath(tenant, realm, path)
	if !exists {
		return nil, false
	}

	// Cache the result
	err := s.cache.Cache(cacheKey, flow, flowCacheTTL, 1)
	if err != nil {
		log.Info().Err(err).Msg("failed to cache flow by path")
	}

	return flow, true
}

// Direct pass-through methods (no caching)
func (s *cachedFlowService) ListFlows(tenant, realm string) ([]model.Flow, error) {
	return s.flowService.ListFlows(tenant, realm)
}

func (s *cachedFlowService) ListAllFlows() ([]model.Flow, error) {
	return s.flowService.ListAllFlows()
}

func (s *cachedFlowService) CreateFlow(tenant, realm string, flow model.Flow) error {
	err := s.flowService.CreateFlow(tenant, realm, flow)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(tenant, realm, flow.Id, flow.Route)
	return nil
}

func (s *cachedFlowService) UpdateFlow(tenant, realm string, flow model.Flow) error {
	// Get the original flow first to get its path for cache invalidation
	originalFlow, exists := s.flowService.GetFlowById(tenant, realm, flow.Id)
	if !exists {
		return fmt.Errorf("flow with id %s not found", flow.Id)
	}

	err := s.flowService.UpdateFlow(tenant, realm, flow)
	if err != nil {
		return err
	}

	// Invalidate both old and new paths
	s.invalidateCaches(tenant, realm, flow.Id, originalFlow.Route) // Invalidate old path
	s.invalidateCaches(tenant, realm, flow.Id, flow.Route)         // Invalidate new path
	return nil
}

func (s *cachedFlowService) DeleteFlow(tenant, realm, id string) error {
	// Get the flow first to get its route for cache invalidation
	flow, exists := s.flowService.GetFlowById(tenant, realm, id)
	if !exists {
		return fmt.Errorf("flow with id %s not found", id)
	}

	err := s.flowService.DeleteFlow(tenant, realm, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(tenant, realm, id, flow.Route)
	return nil
}

func (s *cachedFlowService) ValidateFlowDefinition(content string) ([]FlowLintError, error) {
	return s.flowService.ValidateFlowDefinition(content)
}

// invalidateCaches invalidates all relevant cache entries
func (s *cachedFlowService) invalidateCaches(tenant, realm, id, path string) {
	// Invalidate by ID cache
	idKey := s.getFlowByIdCacheKey(tenant, realm, id)
	s.cache.Invalidate(idKey)

	// Invalidate by path cache
	pathKey := s.getFlowByPathCacheKey(tenant, realm, path)
	s.cache.Invalidate(pathKey)
}
