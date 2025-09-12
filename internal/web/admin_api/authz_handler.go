package admin_api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Identityplane/GoAM/internal/config"
	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

	services_interface "github.com/Identityplane/GoAM/pkg/services"

	"github.com/valyala/fasthttp"
)

type AuthzResponse struct {
	User         AuthzUser                             `json:"user"`
	Entitlements []services_interface.AuthzEntitlement `json:"entitlements"`
}

type AuthzUser struct {
	ID            string    `json:"id"`
	DisplayName   string    `json:"display_name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	GivenName     string    `json:"given_name"`
	FamilyName    string    `json:"family_name"`
	LastLoginAt   time.Time `json:"last_login_at"`
	Locale        string    `json:"locale"`
	Phone         string    `json:"phone"`
	PhoneVerified bool      `json:"phone_verified"`
}

func getUser(ctx *fasthttp.RequestCtx) *model.User {
	userAny := ctx.UserValue("user")
	if userAny == nil {
		return nil
	}
	return userAny.(*model.User)
}

func HandleWhoAmI(ctx *fasthttp.RequestCtx) {
	user := getUser(ctx)

	if user == nil {
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		ctx.SetBodyString("Unauthorized")
		return
	}

	services := service.GetServices()
	entitlements := services.AdminAuthzService.GetEntitlements(user)

	lastLoginAt := time.Time{}
	if user.LastLoginAt != nil {
		lastLoginAt = *user.LastLoginAt
	}

	response := AuthzResponse{
		User: AuthzUser{
			ID:          user.ID,
			LastLoginAt: lastLoginAt,
		},
		Entitlements: entitlements,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal server error")
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonBytes)
}

// ShortTenantInfo represents a tenant and its realms in the admin API
type ShortTenantInfo struct {
	Tenant     string           `json:"tenant"`
	TenantName string           `json:"tenant_name"`
	Realms     []ShortRealmInfo `json:"realms"`
}

// ShortRealmInfo represents a realm in the admin API
type ShortRealmInfo struct {
	Realm     string `json:"realm"`
	RealmName string `json:"realm_name"`
}

// HandleListRealms returns a list of all tenant/realm combinations
// @Summary List all tenant/realm combinations
// @Description Returns a list of all available tenant/realm combinations
// @Tags Realms
// @Accept json
// @Produce json
// @Success 200 {array} ShortTenantInfo
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/realms [get]
func HandleListRealms(ctx *fasthttp.RequestCtx) {
	// Get all loaded realms
	services := service.GetServices()

	userAny := ctx.UserValue("user")
	tenants := make(map[string]*ShortTenantInfo)
	tenantList := make([]ShortTenantInfo, 0, len(tenants))

	var visibleRealms map[string]*services_interface.LoadedRealm
	var err error

	if userAny == nil {

		if config.ServerSettings.UnsafeDisableAdminAuth {
			// if we explicitly disable the authz check we show all realms
			visibleRealms, err = services.RealmService.GetAllRealms()
		} else {
			// Else we return an unauthorized error if we have no user
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString("Unauthorized")
			return
		}
	} else {
		// convert userAny to model.User and get visible realms for user
		user := userAny.(*model.User)
		visibleRealms, err = services.AdminAuthzService.GetVisibleRealms(user)
	}

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to load realms")
		return
	}

	// Group realms by tenant
	tenantList = groupRealmsByTenant(visibleRealms, tenants, tenantList)

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(tenantList, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

func groupRealmsByTenant(visibleRealms map[string]*services_interface.LoadedRealm, tenants map[string]*ShortTenantInfo, tenantList []ShortTenantInfo) []ShortTenantInfo {
	for _, realm := range visibleRealms {

		// Get or create tenant info
		tenant, exists := tenants[realm.Config.Tenant]
		if !exists {
			tenant = &ShortTenantInfo{
				Tenant:     realm.Config.Tenant, // Using ID as label for now, could be enhanced with a mapping
				TenantName: realm.Config.Tenant,
				Realms:     make([]ShortRealmInfo, 0),
			}
			tenants[realm.Config.Tenant] = tenant
		}

		// Add realm to tenant
		tenant.Realms = append(tenant.Realms, ShortRealmInfo{
			Realm:     realm.Config.Realm, // Using ID as label for now, could be enhanced with a mapping
			RealmName: realm.Config.RealmName,
		})
	}

	// Convert map to slice
	for _, tenant := range tenants {
		tenantList = append(tenantList, *tenant)
	}
	return tenantList
}
