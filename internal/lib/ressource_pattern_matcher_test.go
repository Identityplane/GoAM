package lib

import (
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		resource string
		expected bool
		hasError bool
	}{
		// Test cases for /tenant/realm/user/123/ resource
		{
			name:     "should match /tenant/*/*/* pattern",
			pattern:  "/tenant/*/*/*",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match /tenant/realm/user/* pattern",
			pattern:  "/tenant/realm/user/*",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should not match /foobar/*/*/* pattern",
			pattern:  "/foobar/*/*/*",
			resource: "/tenant/realm/user/123",
			expected: false,
			hasError: false,
		},
		{
			name:     "should not match */*/applications/* pattern",
			pattern:  "*/*/applications/*",
			resource: "/tenant/realm/user/123",
			expected: false,
			hasError: false,
		},
		// Additional test cases to demonstrate various pattern features
		{
			name:     "should match exact path",
			pattern:  "/tenant/realm/user/123",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match /**/ pattern for nested directories",
			pattern:  "/tenant/**/user/*",
			resource: "/tenant/realm/subrealm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match single character wildcard",
			pattern:  "/tenant/realm/user/???",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match character class",
			pattern:  "/tenant/realm/user/[0-9]*",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match alternative patterns",
			pattern:  "/tenant/{realm,admin}/user/*",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should not match alternative patterns when not in set",
			pattern:  "/tenant/{admin,internal}/user/*",
			resource: "/tenant/realm/user/123",
			expected: false,
			hasError: false,
		},
		{
			name:     "should match escaped special characters",
			pattern:  "/tenant/realm/user/\\*",
			resource: "/tenant/realm/user/*",
			expected: true,
			hasError: false,
		},
		{
			name:     "should handle empty pattern",
			pattern:  "",
			resource: "/tenant/realm/user/123",
			expected: false,
			hasError: false,
		},
		{
			name:     "should handle empty resource",
			pattern:  "/tenant/*/*/*",
			resource: "",
			expected: false,
			hasError: false,
		},
		{
			name:     "should match root path",
			pattern:  "/*",
			resource: "/tenant",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match deep nested path with **",
			pattern:  "/tenant/**/user/*",
			resource: "/tenant/realm/subrealm/deep/nested/user/123",
			expected: true,
			hasError: false,
		},
		// Additional basic pattern tests
		{
			name:     "should match simple wildcard",
			pattern:  "/tenant/*",
			resource: "/tenant/realm",
			expected: true,
			hasError: false,
		},
		{
			name:     "should match multiple wildcards",
			pattern:  "/tenant/*/user/*",
			resource: "/tenant/realm/user/123",
			expected: true,
			hasError: false,
		},
		{
			name:     "should not match different path structure",
			pattern:  "/tenant/*/admin/*",
			resource: "/tenant/realm/user/123",
			expected: false,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Match(tt.pattern, tt.resource)

			if tt.hasError {
				if err == nil {
					t.Errorf("Match() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Match() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.resource, result, tt.expected)
			}
		})
	}
}

// TestMatchErrorCases tests specific error cases
func TestMatchErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		resource string
	}{
		{
			name:     "invalid character class - missing closing bracket",
			pattern:  "/tenant/[realm/user/*",
			resource: "/tenant/realm/user/123",
		},
		{
			name:     "invalid character class - empty brackets",
			pattern:  "/tenant/[]/user/*",
			resource: "/tenant/realm/user/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Match(tt.pattern, tt.resource)
			if err == nil {
				t.Errorf("Match(%q, %q) expected error but got none", tt.pattern, tt.resource)
			}
		})
	}
}

// TestMatchEntitlementAttributes tests the MatchEntitlementAttributes function
func TestMatchEntitlementAttributes(t *testing.T) {
	tests := []struct {
		name           string
		attributes     []model.EntitlementSetAttributeValue
		resource       string
		action         string
		expectedResult *EntitlementValidationResult
	}{
		{
			name: "deny takes precedence over allow",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow read",
							Resource:    "/tenant/*/user/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
						{
							Description: "Deny read",
							Resource:    "/tenant/*/user/*",
							Action:      "read",
							Effect:      model.EffectTypeDeny,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "Entitlement Deny read has effect deny on resource /tenant/realm/user/123 and action read",
				Allowed:      false,
				denied:       true,
			},
		},
		{
			name: "no deny, at least one allow - should allow",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow read",
							Resource:    "/tenant/*/user/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
						{
							Description: "Allow write",
							Resource:    "/tenant/*/admin/*",
							Action:      "write",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "Resource pattern /tenant/*/user/* and action pattern read match: /tenant/realm/user/123",
				Allowed:      true,
			},
		},
		{
			name: "no matches - should deny",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow read",
							Resource:    "/tenant/*/admin/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
						{
							Description: "Allow write",
							Resource:    "/tenant/*/user/*",
							Action:      "write",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "No entitlement matches",
				Allowed:      false,
			},
		},
		{
			name: "multiple attribute sets - deny in first set takes precedence",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Deny read",
							Resource:    "/tenant/*/user/*",
							Action:      "read",
							Effect:      model.EffectTypeDeny,
						},
					},
				},
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow read",
							Resource:    "/tenant/*/user/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "Entitlement Deny read has effect deny on resource /tenant/realm/user/123 and action read",
				Allowed:      false,
			},
		},
		{
			name: "multiple attribute sets - allow in second set when no deny",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow admin read",
							Resource:    "/tenant/*/admin/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow user read",
							Resource:    "/tenant/*/user/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "Resource pattern /tenant/*/user/* and action pattern read match: /tenant/realm/user/123",
				Allowed:      true,
			},
		},
		{
			name: "empty attributes - should deny",
			attributes: []model.EntitlementSetAttributeValue{},
			resource:   "/tenant/realm/user/123",
			action:     "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "No entitlement matches",
				Allowed:      false,
			},
		},
		{
			name: "empty entitlements in attribute - should deny",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "No entitlement matches",
				Allowed:      false,
			},
		},
		{
			name: "resource pattern mismatch - should deny",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow read",
							Resource:    "/tenant/*/admin/*",
							Action:      "read",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "No entitlement matches",
				Allowed:      false,
			},
		},
		{
			name: "action pattern mismatch - should deny",
			attributes: []model.EntitlementSetAttributeValue{
				{
					Entitlements: []model.Entitlement{
						{
							Description: "Allow write",
							Resource:    "/tenant/*/user/*",
							Action:      "write",
							Effect:      model.EffectTypeAllow,
						},
					},
				},
			},
			resource: "/tenant/realm/user/123",
			action:   "read",
			expectedResult: &EntitlementValidationResult{
				Explentation: "No entitlement matches",
				Allowed:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchEntitlementAttributes(tt.attributes, tt.resource, tt.action)

			if result.Allowed != tt.expectedResult.Allowed {
				t.Errorf("MatchEntitlementAttributes() Allowed = %v, want %v", result.Allowed, tt.expectedResult.Allowed)
			}

			if result.Explentation != tt.expectedResult.Explentation {
				t.Errorf("MatchEntitlementAttributes() Explentation = %q, want %q", result.Explentation, tt.expectedResult.Explentation)
			}

			// Check denied field if expected
			if tt.expectedResult.denied {
				if !result.denied {
					t.Errorf("MatchEntitlementAttributes() expected denied = true, got false")
				}
			}
		})
	}
}

// TestMatchEntitlementAttributesIntegration tests integration with the underlying Match function
func TestMatchEntitlementAttributesIntegration(t *testing.T) {
	// Test that the pattern matching actually works correctly
	attributes := []model.EntitlementSetAttributeValue{
		{
			Entitlements: []model.Entitlement{
				{
					Description: "Allow read on specific user",
					Resource:    "/tenant/realm/user/123",
					Action:      "read",
					Effect:      model.EffectTypeAllow,
				},
				{
					Description: "Deny read on all users",
					Resource:    "/tenant/*/user/*",
					Action:      "read",
					Effect:      model.EffectTypeDeny,
				},
			},
		},
	}

	// This should match the specific allow but be overridden by the deny
	result := MatchEntitlementAttributes(attributes, "/tenant/realm/user/123", "read")
	if result.Allowed {
		t.Errorf("Expected deny to take precedence over allow, but got Allowed = true")
	}
	if !result.denied {
		t.Errorf("Expected denied = true, but got false")
	}
}
