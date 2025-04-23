package admin_api

import (
	"encoding/json"
	"net/http"

	"goiam/internal/service"

	"github.com/valyala/fasthttp"
)

// TenantInfo represents a tenant and its realms in the admin API
type TenantInfo struct {
	Label  string      `json:"label"`
	Value  string      `json:"value"`
	Realms []RealmInfo `json:"realms"`
}

// RealmInfo represents a realm in the admin API
type RealmInfo struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// HandleListRealms returns a list of all tenant/realm combinations
// @Summary List all tenant/realm combinations
// @Description Returns a list of all available tenant/realm combinations
// @Tags tenants
// @Accept json
// @Produce json
// @Success 200 {array} TenantInfo
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/realms [get]
func (h *Handler) HandleListRealms(ctx *fasthttp.RequestCtx) {
	// Get all loaded realms
	services := service.GetServices()
	allRealms := services.RealmService.GetAllRealms()

	// Group realms by tenant
	tenants := make(map[string]*TenantInfo)
	for _, realm := range allRealms {
		tenantID := realm.Config.Tenant
		realmID := realm.Config.Realm

		// Get or create tenant info
		tenant, exists := tenants[tenantID]
		if !exists {
			tenant = &TenantInfo{
				Label:  tenantID, // Using ID as label for now, could be enhanced with a mapping
				Value:  tenantID,
				Realms: make([]RealmInfo, 0),
			}
			tenants[tenantID] = tenant
		}

		// Add realm to tenant
		tenant.Realms = append(tenant.Realms, RealmInfo{
			Label: realmID, // Using ID as label for now, could be enhanced with a mapping
			Value: realmID,
		})
	}

	// Convert map to slice
	tenantList := make([]TenantInfo, 0, len(tenants))
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
