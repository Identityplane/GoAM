package lib

import (
	"fmt"

	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/bmatcuk/doublestar"
)

// Match matches a ressource pattern against a ressource.
// This can be used in various places to check if a user has access to a resource.
func Match(pattern, resource string) (bool, error) {

	match, err := doublestar.Match(pattern, resource)
	if err != nil {
		return false, err
	}

	return match, nil
}

type EntitlementValidationResult struct {
	Explentation string
	Allowed      bool
	denied       bool
}

// Ensures that both the resource and action match the patterns in the entitlement.
// Only allows the entitlement if the effect is allow and the resource and action match the patterns.
func MatchEntitlement(entitlement model.Entitlement, resource, action string) *EntitlementValidationResult {

	// If ressource does not match, return false
	match, err := Match(entitlement.Resource, resource)
	if err != nil {
		return &EntitlementValidationResult{
			Explentation: "Does not match resource pattern",
			Allowed:      false,
		}
	}
	if !match {
		return &EntitlementValidationResult{
			Explentation: "Does not match resource pattern",
			Allowed:      false,
		}
	}

	// If action does not match, return false
	match, err = Match(entitlement.Action, action)
	if err != nil {
		return &EntitlementValidationResult{
			Explentation: "Does not match action pattern",
			Allowed:      false,
		}
	}
	if !match {
		return &EntitlementValidationResult{
			Explentation: "Does not match action pattern",
			Allowed:      false,
		}
	}

	if entitlement.Effect == model.EffectTypeDeny {
		return &EntitlementValidationResult{
			Explentation: fmt.Sprintf("Entitlement %s has effect deny on resource %s and action %s", entitlement.Description, resource, action),
			Allowed:      false,
			denied:       true,
		}
	}

	return &EntitlementValidationResult{
		Explentation: fmt.Sprintf("Resource pattern %s and action pattern %s match: %s", entitlement.Resource, entitlement.Action, resource),
		Allowed:      true,
	}
}

// Check if the ressource and action matches in a list of entitlements.
// If there is a matching entitlement with the effect deny, the result is false.
// Only if there is no deny and at least one allow, the result is true.
func MatchEntitlements(entitlements []model.Entitlement, resource, action string) *EntitlementValidationResult {

	denied := false
	deniedExplentation := ""

	matched := false
	matchedExplentation := ""
	// Check if the ressource and action matches any of the entitlements.
	for _, entitlement := range entitlements {
		match := MatchEntitlement(entitlement, resource, action)
		if match.denied {
			denied = true
			deniedExplentation = match.Explentation
			break
		}
		if match.Allowed {
			matched = true
			matchedExplentation = match.Explentation
		}
	}

	// If it has been denied, return the denied explentation
	if denied {
		return &EntitlementValidationResult{
			Explentation: deniedExplentation,
			Allowed:      false,
			denied:       true,
		}
	}

	// If it has been matched, return the matched explentation
	if matched {
		return &EntitlementValidationResult{
			Explentation: matchedExplentation,
			Allowed:      true,
		}
	}

	// If no entitlement matches, return false
	return &EntitlementValidationResult{
		Explentation: "No entitlement matches",
		Allowed:      false,
	}
}

// Check if the ressource and action matches any of the entitlement attributes.
func MatchEntitlementAttributes(attributes []model.EntitlementSetAttributeValue, resource, action string) *EntitlementValidationResult {

	denied := false
	deniedExplentation := ""

	matched := false
	matchedExplentation := ""

	for _, attribute := range attributes {
		match := MatchEntitlements(attribute.Entitlements, resource, action)
		if match.denied {
			denied = true
			deniedExplentation = match.Explentation
		}
		if match.Allowed {
			matched = true
			matchedExplentation = match.Explentation
		}
	}

	if denied {
		return &EntitlementValidationResult{
			Explentation: deniedExplentation,
			Allowed:      false,
			denied:       true,
		}
	}

	if matched {
		return &EntitlementValidationResult{
			Explentation: matchedExplentation,
			Allowed:      true,
		}
	}

	return &EntitlementValidationResult{
		Explentation: "No entitlement matches",
		Allowed:      false,
	}
}
