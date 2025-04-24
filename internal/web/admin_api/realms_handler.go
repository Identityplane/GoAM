package admin_api

import (
	"encoding/json"
	"net/http"

	"goiam/internal/model"
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
// @Summary Get a specific realm
// @Description Returns the configuration for a specific realm
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
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Realm not found",
		})
		return
	}

	realmConfig := loadedRealm.Config

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(realmConfig, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleCreateRealm creates a new realm
// @Summary Create a new realm
// @Description Creates a new realm with the given configuration
// @Tags Realms
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param request body object{realm_name=string} true "Realm creation payload"
// @Success 201 {object} model.Realm
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/ [post]
func HandleCreateRealm(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	var realmConfig model.Realm
	if err := json.Unmarshal(ctx.PostBody(), &realmConfig); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid request body: " + err.Error())
		return
	}

	if realmConfig.Tenant == "" {
		realmConfig.Tenant = tenant
	}
	if realmConfig.Realm == "" {
		realmConfig.Realm = realm
	}

	// Validate that path parameters match request body
	if realmConfig.Tenant != tenant || realmConfig.Realm != realm {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Path parameters must match request body")
		return
	}

	// Create realm
	if err := service.GetServices().RealmService.CreateRealm(&realmConfig); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to create realm: " + err.Error())
		return
	}

	// Return created realm
	jsonData, err := json.MarshalIndent(realmConfig, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleUpdateRealm updates an existing realm
// @Summary Update a realm
// @Description Updates an existing realm's name
// @Tags Realms
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param request body object{realm_name=string} true "Realm update payload"
// @Success 200 {object} model.Realm
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Realm not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/ [patch]
func HandleUpdateRealm(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	var updatePayload struct {
		RealmName string `json:"realm_name"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &updatePayload); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid request body: " + err.Error())
		return
	}

	// Create realm config with only the updated field
	realmConfig := &model.Realm{
		Tenant:    tenant,
		Realm:     realm,
		RealmName: updatePayload.RealmName,
	}

	// Update realm
	if err := service.GetServices().RealmService.UpdateRealm(realmConfig); err != nil {
		if err.Error() == "realm not found" {
			ctx.SetStatusCode(http.StatusNotFound)
			ctx.SetBodyString("Realm not found")
			return
		}
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to update realm: " + err.Error())
		return
	}

	// Return updated realm
	jsonData, err := json.MarshalIndent(realmConfig, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleDeleteRealm deletes a realm
// @Summary Delete a realm
// @Description Deletes an existing realm
// @Tags Realms
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 204
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/ [delete]
func HandleDeleteRealm(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	if err := service.GetServices().RealmService.DeleteRealm(tenant, realm); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to delete realm: " + err.Error())
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}
