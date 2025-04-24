package admin_api

import (
	"encoding/json"
	"net/http"

	"goiam/internal/service"

	"github.com/valyala/fasthttp"
)

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
	allRealms, err := services.RealmService.GetAllRealms()

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to load realms")
		return
	}

	// Group realms by tenant
	tenants := make(map[string]*ShortTenantInfo)
	for _, realm := range allRealms {

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
	tenantList := make([]ShortTenantInfo, 0, len(tenants))
	for _, tenant := range tenants {
		tenantList = append(tenantList, *tenant)
	}

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

// HandleGetRealm returns a realm
// @Summary List all tenant/realm combinations
// @Description Returns a list of all available tenant/realm combinations
// @Tags Realms
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} model.Realm
// @Failure 404 {string} string "Realm not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/ [get]
func HandleGetRealm(ctx *fasthttp.RequestCtx) {

	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	realmConfig := loadedRealm.Config

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(realmConfig, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
