package service

import (
	"goiam/internal/model"
	"strings"
)

type AuthzEntitlement struct {
	Tenant string   `json:"tenant"`
	Realm  string   `json:"realm"`
	Scopes []string `json:"scopes"`
}

type AdminAuthzService interface {
	GetEntitlements(user *model.User) []AuthzEntitlement
	CheckAccess(user *model.User, tenant, realm, scope string) (bool, string)
	GetVisibleRealms(user *model.User) (map[string]*LoadedRealm, error)
}

type adminAuthzServiceImpl struct {
}

func NewAdminAuthzService() AdminAuthzService {
	return &adminAuthzServiceImpl{}
}

func (s *adminAuthzServiceImpl) GetEntitlements(user *model.User) []AuthzEntitlement {
	entitlements := []AuthzEntitlement{}

	for _, entitlementStr := range user.Entitlements {

		// Split the entitlement string into parts
		parts := strings.SplitN(entitlementStr, "/", 3)
		if len(parts) != 3 {
			continue // Skip invalid entitlement format
		}

		tenant := parts[0]
		realm := parts[1]
		scopes := strings.Fields(parts[2]) // Split scopes by whitespace

		entitlements = append(entitlements, AuthzEntitlement{
			Tenant: tenant,
			Realm:  realm,
			Scopes: scopes,
		})
	}

	return entitlements
}

func (s *adminAuthzServiceImpl) GetVisibleRealms(user *model.User) (map[string]*LoadedRealm, error) {

	// Load all realms via realm service
	services := GetServices()
	realms, err := services.RealmService.GetAllRealms()
	if err != nil {
		return nil, err
	}

	// For each realm, check if the user has access to it and return only the realm with access
	visibleRealms := make(map[string]*LoadedRealm)
	for realmId, realm := range realms {
		hasAccess, _ := s.CheckAccess(user, realm.Config.Tenant, realm.Config.Realm, "")
		if hasAccess {
			visibleRealms[realmId] = realm
		}
	}

	// Return the results
	return visibleRealms, nil
}

// CheckAccess checks if a user has access to a specific tenant, realm, and scope.
// Returns true if access is granted, false otherwise, and the matching entitlement string if access is granted.
// If scope is empty, it checks if the user has any access to the specified tenant/realm.
func (s *adminAuthzServiceImpl) CheckAccess(user *model.User, tenant, realm, scope string) (bool, string) {
	if user == nil || len(user.Entitlements) == 0 {
		return false, ""
	}

	for _, entitlementStr := range user.Entitlements {
		parts := strings.SplitN(entitlementStr, "/", 3)
		if len(parts) != 3 {
			continue // Skip invalid entitlement format
		}

		entTenant := parts[0]
		entRealm := parts[1]
		entScopes := strings.Fields(parts[2])

		// Check tenant match (exact or wildcard)
		if entTenant != "*" && entTenant != tenant {
			continue
		}

		// Check realm match (exact or wildcard)
		if entRealm != "*" && entRealm != realm {
			continue
		}

		// If scope is empty, we just need to check if there are any scopes
		if scope == "" {
			if len(entScopes) > 0 {
				return true, entitlementStr
			}
			continue
		}

		// Check scope match (exact or wildcard)
		for _, entScope := range entScopes {
			if entScope == "*" || entScope == scope {
				return true, entitlementStr
			}
		}
	}

	return false, ""
}
