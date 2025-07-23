package admin_api

import (
	"encoding/json"
	"net/http"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/valyala/fasthttp"
)

// HandleListApplications returns all applications for a realm
// @Summary List all applications
// @Description Returns a list of all applications in a realm
// @Tags Applications
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {array} model.Application
// @Failure 404 {string} string "Realm not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/applications [get]
func HandleListApplications(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	apps, err := service.GetServices().ApplicationService.ListApplications(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to list applications: " + err.Error(),
		})
		return
	}

	// Ensure apps is never nil
	if apps == nil {
		apps = []model.Application{}
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(apps, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleGetApplication returns a specific application
// @Summary Get an application
// @Description Returns a specific application configuration
// @Tags Applications
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param client_id path string true "Client ID"
// @Success 200 {object} model.Application
// @Failure 404 {string} string "Application not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/applications/{client_id} [get]
func HandleGetApplication(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	clientId := ctx.UserValue("client_id").(string)

	app, ok := service.GetServices().ApplicationService.GetApplication(tenant, realm, clientId)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Application not found",
		})
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleCreateApplication creates a new application
// @Summary Create an application
// @Description Creates a new application configuration
// @Tags Applications
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param client_id path string true "Client ID"
// @Param request body model.Application true "Application configuration"
// @Success 201 {object} model.Application
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/applications/{client_id} [post]
func HandleCreateApplication(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	clientId := ctx.UserValue("client_id").(string)

	// Parse JSON request body
	var app model.Application
	if err := json.Unmarshal(ctx.PostBody(), &app); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid JSON content: " + err.Error(),
		})
		return
	}

	// Set tenant, realm and client_id from URL parameters
	app.Tenant = tenant
	app.Realm = realm
	app.ClientId = clientId

	if err := service.GetServices().ApplicationService.CreateApplication(tenant, realm, app); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to create application: " + err.Error(),
		})
		return
	}

	// Return created application
	jsonData, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleUpdateApplication updates an existing application
// @Summary Update an application
// @Description Updates an existing application configuration
// @Tags Applications
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param client_id path string true "Client ID"
// @Param request body model.Application true "Application configuration"
// @Success 200 {object} model.Application
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Application not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/applications/{client_id} [put]
func HandleUpdateApplication(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	clientId := ctx.UserValue("client_id").(string)

	// Parse JSON request body
	var app model.Application
	if err := json.Unmarshal(ctx.PostBody(), &app); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid JSON content: " + err.Error(),
		})
		return
	}

	// Set tenant, realm and client_id from URL parameters
	app.Tenant = tenant
	app.Realm = realm
	app.ClientId = clientId

	if err := service.GetServices().ApplicationService.UpdateApplication(tenant, realm, app); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to update application: " + err.Error(),
		})
		return
	}

	// Return updated application
	jsonData, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleDeleteApplication deletes an application
// @Summary Delete an application
// @Description Deletes an existing application
// @Tags Applications
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param client_id path string true "Client ID"
// @Success 204
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/applications/{client_id} [delete]
func HandleDeleteApplication(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	clientId := ctx.UserValue("client_id").(string)

	if err := service.GetServices().ApplicationService.DeleteApplication(tenant, realm, clientId); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to delete application: " + err.Error(),
		})
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}

// HandleRegenerateClientSecret generates a new client secret for an application
// @Summary Regenerate client secret
// @Description Generates a new client secret for an application
// @Tags Applications
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param client_id path string true "Client ID"
// @Success 200 {object} map[string]string "Client secret"
// @Failure 404 {string} string "Application not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/applications/{client_id}/regenerate-secret [post]
func HandleRegenerateClientSecret(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	clientId := ctx.UserValue("client_id").(string)

	clientSecret, err := service.GetServices().ApplicationService.RegenerateClientSecret(tenant, realm, clientId)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to regenerate client secret: " + err.Error(),
		})
		return
	}

	// Return the new client secret
	jsonData, err := json.MarshalIndent(map[string]string{
		"client_secret": clientSecret,
	}, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
