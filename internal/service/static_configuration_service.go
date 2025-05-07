package service

import (
	"fmt"
	"goiam/internal/logger"
	"goiam/internal/model"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Represents the static configuration for a realm as yaml
type realmYaml struct {
	Realm        string                        `yaml:"realm"`
	RealmName    string                        `yaml:"realm_name"`
	Tenant       string                        `yaml:"tenant"`
	BaseUrl      string                        `yaml:"base_url"`
	Applications map[string]*model.Application `yaml:"applications"`
	Flows        map[string]*model.Flow        `yaml:"flows"`
}

type StaticConfigurationService interface {
	LoadConfigurationFromFiles(configRoot string) error
}

type staticConfigurationServiceImpl struct {
}

func NewStaticConfigurationService() StaticConfigurationService {
	return &staticConfigurationServiceImpl{}
}

func (s *staticConfigurationServiceImpl) LoadConfigurationFromFiles(configRoot string) error {

	var realmService RealmService = GetServices().RealmService
	var flowService FlowService = GetServices().FlowService
	var applicationService ApplicationService = GetServices().ApplicationService

	// Load realms from config directory
	realmsFromConfigDir, err := loadRealmsFromConfigDir(configRoot)
	if err != nil {
		return fmt.Errorf("failed to load realms from config directory %s: %w", configRoot, err)
	}

	// Create new realms, flows and applications using the other services
	for _, realm := range realmsFromConfigDir {

		// If realm does not already exist, create it
		_, exists := realmService.GetRealm(realm.Tenant, realm.Realm)
		if !exists {
			logger.DebugNoContext("Creating realm: %s", realm.Realm)
			err := realmService.CreateRealm(&model.Realm{
				Realm:     realm.Realm,
				RealmName: realm.RealmName,
				Tenant:    realm.Tenant,
				BaseUrl:   realm.BaseUrl,
			})
			if err != nil {
				logger.PanicNoContext("Failed to create realm %s: %s", realm.Realm, err)
			}
		}

		// Create flows of not existing
		for _, flow := range realm.Flows {
			_, exists := flowService.GetFlowById(realm.Tenant, realm.Realm, flow.Id)
			if !exists {
				logger.DebugNoContext("Creating flow: %s", flow.Id)
				err := flowService.CreateFlow(realm.Tenant, realm.Realm, *flow)
				if err != nil {
					logger.PanicNoContext("Failed to create flow %s: %s", flow.Id, err)
				}
			}
		}

		// Create applications of not existing
		for _, application := range realm.Applications {
			_, exists := applicationService.GetApplication(realm.Tenant, realm.Realm, application.ClientId)
			if !exists {
				logger.DebugNoContext("Creating application: %s", application.ClientId)
				err := applicationService.CreateApplication(realm.Tenant, realm.Realm, *application)
				if err != nil {
					logger.PanicNoContext("Failed to create application %s: %s", application.ClientId, err)
				}
			}
		}
	}

	return nil
}

// loadRealmsFromConfigDir loads all realm configurations from the given config root directory
func loadRealmsFromConfigDir(configRoot string) ([]realmYaml, error) {

	var newRealms []realmYaml

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
		if currentDepth != 2 {
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

		newRealms = append(newRealms, *cfg)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk realm config directory %s: %w", configRoot, err)
	}

	logger.DebugNoContext("Loaded %d realms from config directory", len(newRealms))
	return newRealms, nil

}

// Helper function to load realm config from file
func loadRealmConfigFromFilePath(path string) (*realmYaml, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var yamlConfig realmYaml
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
	}

	if yamlConfig.Realm == "" {
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
	realm := yamlConfig.Realm

	// Ensure for all flows and appliactions the tenant and realm are set correctly
	// and necessary fields are set
	yamlConfig.Tenant = tenant
	yamlConfig.Realm = realm

	for id, flow := range yamlConfig.Flows {
		flow.Tenant = tenant
		flow.Realm = realm
		flow.Id = id
	}
	for id, app := range yamlConfig.Applications {
		app.Tenant = tenant
		app.Realm = realm
		app.ClientId = id
	}

	// For each flow, we need to read the flow definition yaml
	for id, flow := range yamlConfig.Flows {
		flowDef, err := os.ReadFile(filepath.Join(filepath.Dir(path), realm, "flows", flow.DefinitionLocation))
		if err != nil {
			return nil, fmt.Errorf("failed to read flow definition for %s: %w", id, err)
		}
		flow.DefinitionLocation = filepath.Join(filepath.Dir(path), "flows", id+".yaml")
		flow.DefinitionYaml = string(flowDef)
	}

	return &yamlConfig, nil
}
