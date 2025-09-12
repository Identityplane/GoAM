package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

type adminAuthzServiceImpl struct {
}

func NewAdminAuthzService() services_interface.AdminAuthzService {
	return &adminAuthzServiceImpl{}
}

func (s *adminAuthzServiceImpl) GetEntitlements(user *model.User) []services_interface.AuthzEntitlement {

	entitlements := []services_interface.AuthzEntitlement{}

	entitlementsAttrs, _, err := model.GetAttributes[model.EntitlementSetAttributeValue](user, model.AttributeTypeEntitlements)
	if err != nil {
		return nil
	}

	for _, entitlementAttr := range entitlementsAttrs {
		for _, entitlement := range entitlementAttr.Entitlements {

			entitlements = append(entitlements, services_interface.AuthzEntitlement{
				Resource: entitlement.Resource,
				Action:   entitlement.Action,
				Effect:   string(entitlement.Effect),
			})
		}
	}

	return entitlements
}

func (s *adminAuthzServiceImpl) GetVisibleRealms(user *model.User) (map[string]*services_interface.LoadedRealm, error) {

	// Load all realms via realm service
	services := GetServices()
	realms, err := services.RealmService.GetAllRealms()
	if err != nil {
		return nil, err
	}

	// For each realm, check if the user has access to it and return only the realm with access
	visibleRealms := make(map[string]*services_interface.LoadedRealm)
	for realmId, realm := range realms {

		resource := fmt.Sprintf("%s/%s", realm.Config.Tenant, realm.Config.Realm)
		hasAccess, _ := s.CheckAccess(user, resource, "GET")
		if hasAccess {
			visibleRealms[realmId] = realm
		}
	}

	// Return the results
	return visibleRealms, nil
}

// CheckAccess checks if a 	user has access to a specific tenant, realm, and scope.
// Returns true if access is granted, false otherwise, and the matching entitlement string if access is granted.
// If scope is empty, it checks if the user has any access to the specified tenant/realm.
// Returns the result and the explanation for the result
func (s *adminAuthzServiceImpl) CheckAccess(user *model.User, resource, action string) (bool, string) {
	if user == nil {
		return false, ""
	}

	entitlements, _, err := model.GetAttributes[model.EntitlementSetAttributeValue](user, model.AttributeTypeEntitlements)
	if err != nil {
		return false, ""
	}

	// Sue the entitlements against the resource and action
	result := lib.MatchEntitlementAttributes(entitlements, resource, action)

	return result.Allowed, result.Explentation
}

func (s *adminAuthzServiceImpl) CreateTenant(tenantSlug, tenantName string, user *model.User) error {

	services := GetServices()

	// Validate tenant slug
	if len(tenantSlug) < 3 {
		return fmt.Errorf("tenant slug must be at least 3 characters")
	}

	if len(tenantSlug) > 50 {
		return fmt.Errorf("tenant slug must be less than 50 characters")
	}

	matched, err := regexp.MatchString("^[a-z0-9-]+$", tenantSlug)
	if err != nil {
		return fmt.Errorf("error validating tenant slug: %v", err)
	}
	if !matched {
		return fmt.Errorf("tenant slug can only contain lowercase letters, numbers, and hyphens")
	}

	// add an entitlement to the user
	entitlementAttr := model.UserAttribute{
		Tenant: "internal",
		Realm:  "internal",
		UserID: user.ID,
		Type:   model.AttributeTypeEntitlements,
		Value: &model.EntitlementSetAttributeValue{
			Entitlements: []model.Entitlement{
				{
					Description: fmt.Sprintf("Creator of tenant '%s'", tenantName),
					Resource:    fmt.Sprintf("%s/**", tenantSlug),
					Action:      "*",
					Effect:      model.EffectTypeAllow,
				},
			},
		},
	}

	// Save the entitlement to the user
	services.UserAttributeService.CreateUserAttribute(context.Background(), entitlementAttr)

	// create the realm
	err = services.RealmService.CreateRealm(&model.Realm{
		Tenant:    tenantSlug,
		Realm:     "default",
		RealmName: "Default Realm",
	})

	if err != nil {
		return err
	}

	return nil
}
