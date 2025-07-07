package service

import (
	"testing"

	"github.com/gianlucafrei/GoAM/internal/model"
)

func TestGetEntitlements(t *testing.T) {
	authzService := NewAdminAuthzService()

	tests := []struct {
		name        string
		user        *model.User
		want        []AuthzEntitlement
		description string
	}{
		{
			name: "single entitlement",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/customers/user:write flows:read",
				},
			},
			want: []AuthzEntitlement{
				{
					Tenant: "acme",
					Realm:  "customers",
					Scopes: []string{"user:write", "flows:read"},
				},
			},
			description: "should parse a single entitlement with multiple scopes",
		},
		{
			name: "multiple entitlements",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/customers/user:write flows:read",
					"acme/admin/admin:write",
				},
			},
			want: []AuthzEntitlement{
				{
					Tenant: "acme",
					Realm:  "customers",
					Scopes: []string{"user:write", "flows:read"},
				},
				{
					Tenant: "acme",
					Realm:  "admin",
					Scopes: []string{"admin:write"},
				},
			},
			description: "should parse multiple entitlements",
		},
		{
			name: "empty entitlements",
			user: &model.User{
				ID:           "test-user",
				Entitlements: []string{},
			},
			want:        []AuthzEntitlement{},
			description: "should return empty slice for user with no entitlements",
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
				if got[i].Tenant != want.Tenant {
					t.Errorf("GetEntitlements() tenant = %v, want %v", got[i].Tenant, want.Tenant)
				}
				if got[i].Realm != want.Realm {
					t.Errorf("GetEntitlements() realm = %v, want %v", got[i].Realm, want.Realm)
				}
				if len(got[i].Scopes) != len(want.Scopes) {
					t.Errorf("GetEntitlements() scopes length = %v, want %v", len(got[i].Scopes), len(want.Scopes))
					continue
				}
				for j, scope := range want.Scopes {
					if got[i].Scopes[j] != scope {
						t.Errorf("GetEntitlements() scope[%d] = %v, want %v", j, got[i].Scopes[j], scope)
					}
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
		tenant      string
		realm       string
		scope       string
		wantAccess  bool
		wantReason  string
		description string
	}{
		{
			name: "exact match",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/customers/user:write flows:read",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "user:write",
			wantAccess:  true,
			wantReason:  "acme/customers/user:write flows:read",
			description: "should grant access when all parts match exactly",
		},
		{
			name: "empty scope with access",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/customers/user:read",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "",
			wantAccess:  true,
			wantReason:  "acme/customers/user:read",
			description: "should grant access with empty scope when user has any access to realm",
		},
		{
			name: "empty scope without access",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/admin/user:read",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "",
			wantAccess:  false,
			wantReason:  "",
			description: "should deny access with empty scope when user has no access to realm",
		},
		{
			name: "empty scope with wildcard tenant",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"*/customers/user:read",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "",
			wantAccess:  true,
			wantReason:  "*/customers/user:read",
			description: "should grant access with empty scope and wildcard tenant",
		},
		{
			name: "empty scope with wildcard realm",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/*/user:read",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "",
			wantAccess:  true,
			wantReason:  "acme/*/user:read",
			description: "should grant access with empty scope and wildcard realm",
		},
		{
			name: "wildcard tenant",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"*/customers/user:write",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "user:write",
			wantAccess:  true,
			wantReason:  "*/customers/user:write",
			description: "should grant access with wildcard tenant",
		},
		{
			name: "wildcard realm",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/*/user:write",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "user:write",
			wantAccess:  true,
			wantReason:  "acme/*/user:write",
			description: "should grant access with wildcard realm",
		},
		{
			name: "wildcard scope",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/customers/*",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "user:write",
			wantAccess:  true,
			wantReason:  "acme/customers/*",
			description: "should grant access with wildcard scope",
		},
		{
			name: "no match",
			user: &model.User{
				ID: "test-user",
				Entitlements: []string{
					"acme/customers/user:read",
				},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "user:write",
			wantAccess:  false,
			wantReason:  "",
			description: "should deny access when no entitlement matches",
		},
		{
			name: "empty entitlements",
			user: &model.User{
				ID:           "test-user",
				Entitlements: []string{},
			},
			tenant:      "acme",
			realm:       "customers",
			scope:       "user:write",
			wantAccess:  false,
			wantReason:  "",
			description: "should deny access when user has no entitlements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccess, gotReason := authzService.CheckAccess(tt.user, tt.tenant, tt.realm, tt.scope)
			if gotAccess != tt.wantAccess {
				t.Errorf("CheckAccess() access = %v, want %v", gotAccess, tt.wantAccess)
			}
			if gotReason != tt.wantReason {
				t.Errorf("CheckAccess() reason = %v, want %v", gotReason, tt.wantReason)
			}
		})
	}
}
