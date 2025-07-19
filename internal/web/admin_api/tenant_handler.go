package admin_api

import (
	"encoding/json"

	"github.com/Identityplane/GoAM/internal/service"

	"github.com/valyala/fasthttp"
)

// TenantAvailabilityResponse represents the response for tenant name availability check
type TenantAvailabilityResponse struct {
	Available bool `json:"available"`
}

// CreateTenantRequest represents the request body for tenant creation
type CreateTenantRequest struct {
	TenantSlug string `json:"tenant_slug"`
	TenantName string `json:"tenant_name"`
}

// HandleTenantNameAvailable checks if a tenant name is available
// @Summary Check tenant name availability
// @Description Checks if a tenant name is available for registration
// @Tags Tenants
// @Accept json
// @Produce json
// @Param tenant_name path string true "Tenant name to check"
// @Success 200 {object} TenantAvailabilityResponse
// @Failure 500 {string} string "Internal server error"
// @Router /admin/tenants/check-availability/{tenant_name} [get]
func HandleTenantNameAvailable(ctx *fasthttp.RequestCtx) {
	tenantName := ctx.UserValue("tenant_name").(string)

	available, err := service.GetServices().RealmService.IsTenantNameAvailable(tenantName)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		errorResponse := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		jsonData, _ := json.Marshal(errorResponse)
		ctx.SetBody(jsonData)
		return
	}

	response := TenantAvailabilityResponse{
		Available: available,
	}

	ctx.SetContentType("application/json")
	jsonData, err := json.Marshal(response)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal server error")
		return
	}
	ctx.SetBody(jsonData)
}

// HandleCreateTenant creates a new tenant
// @Summary Create a new tenant
// @Description Creates a new tenant with the provided slug and name
// @Tags Tenants
// @Accept json
// @Produce json
// @Param request body CreateTenantRequest true "Tenant creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /admin/tenants [post]
func HandleCreateTenant(ctx *fasthttp.RequestCtx) {
	var req CreateTenantRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("text/plain")
		ctx.SetBodyString("Invalid request body")
		return
	}

	available, err := service.GetServices().RealmService.IsTenantNameAvailable(req.TenantName)
	if err != nil || !available {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("text/plain")
		ctx.SetBodyString("Not available")
		return
	}

	user := getUser(ctx)
	if user == nil {
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		ctx.SetContentType("text/plain")
		ctx.SetBodyString("Unauthorized")
		return
	}

	err = service.GetServices().AdminAuthzService.CreateTenant(req.TenantSlug, req.TenantName, user)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(err.Error())
		return
	}

	response := map[string]interface{}{
		"tenant_slug": req.TenantSlug,
		"tenant_name": req.TenantName,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal server error")
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
