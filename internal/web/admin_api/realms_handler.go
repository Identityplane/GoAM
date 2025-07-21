package admin_api

import (
	"encoding/json"
	"net/http"

	"github.com/Identityplane/GoAM/internal/model"
	"github.com/Identityplane/GoAM/internal/web/webutils"

	"github.com/Identityplane/GoAM/internal/service"

	"github.com/valyala/fasthttp"
)

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

	// For each visible realm set the base url to the fallback url if not set
	if realmConfig.BaseUrl == "" {
		realmConfig.BaseUrl = webutils.GetFallbackUrl(ctx, tenant, realm)
	}

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
