package admin_api

import (
	"encoding/json"
	"net/http"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

	"github.com/valyala/fasthttp"
)

// @Summary List user attributes
// @Description Get all attributes for a user (without detailed values)
// @Tags User Attributes
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Success 200 {array} model.UserAttribute
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id}/attributes [get]
func HandleListUserAttributes(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	userID := ctx.UserValue("id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get user attributes from service
	attributes, err := service.GetServices().UserAttributeService.ListUserAttributes(ctx, tenant, realm, userID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to list user attributes: " + err.Error())
		return
	}

	if attributes == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// If no attributes, return empty array (not nil)
	if len(attributes) == 0 {
		attributes = []*model.UserAttribute{}
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(attributes, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// @Summary Create user attribute
// @Description Create a new attribute for a user
// @Tags User Attributes
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Param attribute body model.UserAttribute true "User attribute object"
// @Success 201 {object} model.UserAttribute
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id}/attributes [post]
func HandleCreateUserAttribute(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	userID := ctx.UserValue("id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Parse request body
	var createAttribute model.UserAttribute
	if err := json.Unmarshal(ctx.PostBody(), &createAttribute); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Set tenant, realm, and user ID from path
	createAttribute.Tenant = tenant
	createAttribute.Realm = realm
	createAttribute.UserID = userID

	// Create attribute through service
	attribute, err := service.GetServices().UserAttributeService.CreateUserAttribute(ctx, createAttribute)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to create user attribute: " + err.Error())
		return
	}

	if attribute == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(attribute, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusCreated)
	ctx.SetBody(jsonData)
}

// @Summary Get user attribute
// @Description Get a specific attribute instance with full details
// @Tags User Attributes
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Param attribute-type path string true "Attribute type"
// @Param attribute-id path string true "Attribute ID"
// @Success 200 {object} model.UserAttribute
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id} [get]
func HandleGetUserAttribute(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	attributeID := ctx.UserValue("attribute-id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get attribute from service
	attribute, err := service.GetServices().UserAttributeService.GetUserAttributeByID(ctx, tenant, realm, attributeID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user attribute: " + err.Error())
		return
	}

	if attribute == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Attribute not found")
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(attribute, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// @Summary Update user attribute
// @Description Update an existing attribute instance
// @Tags User Attributes
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Param attribute-type path string true "Attribute type"
// @Param attribute-id path string true "Attribute ID"
// @Param attribute body model.UserAttribute true "User attribute object"
// @Success 200 {object} model.UserAttribute
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id} [patch]
func HandleUpdateUserAttribute(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	attributeID := ctx.UserValue("attribute-id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Parse request body
	var updateAttribute model.UserAttribute
	if err := json.Unmarshal(ctx.PostBody(), &updateAttribute); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Set path parameters
	updateAttribute.Tenant = tenant
	updateAttribute.Realm = realm
	updateAttribute.ID = attributeID

	// Update attribute through service
	err := service.GetServices().UserAttributeService.UpdateUserAttribute(ctx, &updateAttribute)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to update user attribute: " + err.Error())
		return
	}

	// Get the updated attribute
	attribute, err := service.GetServices().UserAttributeService.GetUserAttributeByID(ctx, tenant, realm, attributeID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get updated user attribute: " + err.Error())
		return
	}

	if attribute == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Attribute not found")
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(attribute, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// @Summary Delete user attribute
// @Description Remove a specific attribute instance
// @Tags User Attributes
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Param attribute-type path string true "Attribute type"
// @Param attribute-id path string true "Attribute ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id} [delete]
func HandleDeleteUserAttribute(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	attributeID := ctx.UserValue("attribute-id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Delete attribute through service
	err := service.GetServices().UserAttributeService.DeleteUserAttribute(ctx, tenant, realm, attributeID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to delete user attribute: " + err.Error())
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}
