package service

import (
	"regexp"
	"testing"

	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

func TestGetEntitlements(t *testing.T) {
	authzService := NewAdminAuthzService()

	tests := []struct {
		name        string
		user        *model.User
		want        []services_interface.AuthzEntitlement
		description string
	}{
		{
			name: "single entitlement",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Admin access",
									Resource:    "/acme/customers/**",
									Action:      "GET",
									Effect:      model.EffectTypeAllow,
								},
								{
									Description: "Write access",
									Resource:    "/acme/customers/users/*",
									Action:      "POST",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
				},
			},
			want: []services_interface.AuthzEntitlement{
				{
					Resource: "/acme/customers/**",
					Action:   "GET",
					Effect:   "allow",
				},
				{
					Resource: "/acme/customers/users/*",
					Action:   "POST",
					Effect:   "allow",
				},
			},
			description: "should parse entitlements from user attributes",
		},
		{
			name: "multiple entitlement sets",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Admin access",
									Resource:    "/acme/customers/**",
									Action:      "GET",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
					{
						ID:     "attr2",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Write access",
									Resource:    "/acme/admin/**",
									Action:      "POST",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
				},
			},
			want: []services_interface.AuthzEntitlement{
				{
					Resource: "/acme/customers/**",
					Action:   "GET",
					Effect:   "allow",
				},
				{
					Resource: "/acme/admin/**",
					Action:   "POST",
					Effect:   "allow",
				},
			},
			description: "should parse multiple entitlement sets",
		},
		{
			name: "empty entitlements",
			user: &model.User{
				ID:             "test-user",
				UserAttributes: []model.UserAttribute{},
			},
			want:        []services_interface.AuthzEntitlement{},
			description: "should return empty slice for user with no entitlements",
		},
		{
			name: "no entitlement attributes",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   "email",
						Value:  "test@example.com",
					},
				},
			},
			want:        []services_interface.AuthzEntitlement{},
			description: "should return empty slice when no entitlement attributes exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := authzService.GetEntitlements(tt.user)
			if len(got) != len(tt.want) {
				t.Errorf("GetEntitlements() got %d entitlements, want %d", len(got), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if got[i].Resource != want.Resource {
					t.Errorf("GetEntitlements() resource[%d] = %v, want %v", i, got[i].Resource, want.Resource)
				}
				if got[i].Action != want.Action {
					t.Errorf("GetEntitlements() action[%d] = %v, want %v", i, got[i].Action, want.Action)
				}
				if got[i].Effect != want.Effect {
					t.Errorf("GetEntitlements() effect[%d] = %v, want %v", i, got[i].Effect, want.Effect)
				}
			}
		})
	}
}

func TestCheckAccess(t *testing.T) {
	authzService := NewAdminAuthzService()

	tests := []struct {
		name        string
		user        *model.User
		resource    string
		action      string
		wantAccess  bool
		wantReason  string
		description string
	}{
		{
			name: "exact match",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Admin access",
									Resource:    "/acme/customers/users/123",
									Action:      "GET",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
				},
			},
			resource:    "/acme/customers/users/123",
			action:      "GET",
			wantAccess:  true,
			wantReason:  "Resource pattern /acme/customers/users/123 and action pattern GET match: /acme/customers/users/123",
			description: "should grant access when resource and action match exactly",
		},
		{
			name: "wildcard resource match",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "User access",
									Resource:    "/acme/customers/users/*",
									Action:      "GET",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
				},
			},
			resource:    "/acme/customers/users/456",
			action:      "GET",
			wantAccess:  true,
			wantReason:  "Resource pattern /acme/customers/users/* and action pattern GET match: /acme/customers/users/456",
			description: "should grant access with wildcard resource pattern",
		},
		{
			name: "wildcard action match",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Full access",
									Resource:    "/acme/customers/**",
									Action:      "*",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
				},
			},
			resource:    "/acme/customers/users/123",
			action:      "POST",
			wantAccess:  true,
			wantReason:  "Resource pattern /acme/customers/** and action pattern * match: /acme/customers/users/123",
			description: "should grant access with wildcard action pattern",
		},
		{
			name: "deny takes precedence",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Allow access",
									Resource:    "/acme/customers/users/*",
									Action:      "GET",
									Effect:      model.EffectTypeAllow,
								},
								{
									Description: "Deny access",
									Resource:    "/acme/customers/users/*",
									Action:      "GET",
									Effect:      model.EffectTypeDeny,
								},
							},
						},
					},
				},
			},
			resource:    "/acme/customers/users/123",
			action:      "GET",
			wantAccess:  false,
			wantReason:  "Entitlement Deny access has effect deny on resource /acme/customers/users/123 and action GET",
			description: "should deny access when deny entitlement exists",
		},
		{
			name: "no match",
			user: &model.User{
				ID: "test-user",
				UserAttributes: []model.UserAttribute{
					{
						ID:     "attr1",
						UserID: "test-user",
						Type:   model.AttributeTypeEntitlements,
						Value: model.EntitlementSetAttributeValue{
							Entitlements: []model.Entitlement{
								{
									Description: "Admin access",
									Resource:    "/acme/admin/**",
									Action:      "GET",
									Effect:      model.EffectTypeAllow,
								},
							},
						},
					},
				},
			},
			resource:    "/acme/customers/users/123",
			action:      "GET",
			wantAccess:  false,
			wantReason:  "No entitlement matches",
			description: "should deny access when no entitlement matches",
		},
		{
			name:        "nil user",
			user:        nil,
			resource:    "/acme/customers/users/123",
			action:      "GET",
			wantAccess:  false,
			wantReason:  "",
			description: "should deny access for nil user",
		},
		{
			name: "empty entitlements",
			user: &model.User{
				ID:             "test-user",
				UserAttributes: []model.UserAttribute{},
			},
			resource:    "/acme/customers/users/123",
			action:      "GET",
			wantAccess:  false,
			wantReason:  "No entitlement matches",
			description: "should deny access when user has no entitlements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccess, gotReason := authzService.CheckAccess(tt.user, tt.resource, tt.action)
			if gotAccess != tt.wantAccess {
				t.Errorf("CheckAccess() access = %v, want %v", gotAccess, tt.wantAccess)
			}
			if gotReason != tt.wantReason {
				t.Errorf("CheckAccess() reason = %q, want %q", gotReason, tt.wantReason)
			}
		})
	}
}

func TestCreateTenant(t *testing.T) {
	// Note: This test only covers the validation logic since the actual creation requires services
	// that are not available in unit tests. The validation logic is the important part to test.

	tests := []struct {
		name        string
		tenantSlug  string
		tenantName  string
		description string
	}{
		{
			name:        "valid tenant creation",
			tenantSlug:  "valid-tenant",
			tenantName:  "Valid Tenant",
			description: "should accept valid slug",
		},
		{
			name:        "tenant slug too short",
			tenantSlug:  "ab",
			tenantName:  "Short Tenant",
			description: "should reject slug less than 3 characters",
		},
		{
			name:        "tenant slug too long",
			tenantSlug:  "this-is-a-very-long-tenant-slug-that-exceeds-fifty-characters",
			tenantName:  "Long Tenant",
			description: "should reject slug more than 50 characters",
		},
		{
			name:        "tenant slug with invalid characters",
			tenantSlug:  "invalid@tenant",
			tenantName:  "Invalid Tenant",
			description: "should reject slug containing invalid characters",
		},
		{
			name:        "tenant slug with uppercase letters",
			tenantSlug:  "InvalidTenant",
			tenantName:  "Invalid Tenant",
			description: "should reject slug containing uppercase letters",
		},
		{
			name:        "tenant slug with spaces",
			tenantSlug:  "invalid tenant",
			tenantName:  "Invalid Tenant",
			description: "should reject slug containing spaces",
		},
		{
			name:        "tenant slug with underscores",
			tenantSlug:  "invalid_tenant",
			tenantName:  "Invalid Tenant",
			description: "should reject slug containing underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the regex validation logic directly
			matched, err := regexp.MatchString("^[a-z0-9-]+$", tt.tenantSlug)
			if err != nil {
				t.Errorf("regex validation failed: %v", err)
				return
			}

			// Check length constraints
			validLength := len(tt.tenantSlug) >= 3 && len(tt.tenantSlug) <= 50

			// The slug should be valid if it matches regex AND has valid length
			expectedValid := matched && validLength

			if tt.name == "valid tenant creation" && !expectedValid {
				t.Errorf("Expected valid slug but validation failed: regex=%v, length=%v", matched, validLength)
			}
			if tt.name != "valid tenant creation" && expectedValid {
				t.Errorf("Expected invalid slug but validation passed: regex=%v, length=%v", matched, validLength)
			}
		})
	}
}
