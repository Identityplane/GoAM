package web

import (
	"encoding/json"
	"goiam/internal/db/model"
	"goiam/internal/db/service"
	"goiam/internal/realms"
	"net/http"
	"strconv"

	"github.com/valyala/fasthttp"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	userService service.UserAdminService
}

// NewAdminHandler creates a new AdminHandler instance
func NewAdminHandler(userService service.UserAdminService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
	}
}

// HandleListUsers handles the GET /admin/users endpoint
func (h *AdminHandler) HandleListUsers(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	page, err1 := ctx.QueryArgs().GetUint("page")
	pageSize, err2 := ctx.QueryArgs().GetUint("page_size")

	// If no pagination is provided, set default values
	if err1 != nil || err2 != nil {
		page = 1
		pageSize = 100
	}

	pagination := service.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// List users from the service
	users, total, err := h.userService.ListUsers(ctx, tenant, realm, pagination)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to list users")
		return
	}

	// Set total count and pagination to response headers
	ctx.Response.Header.Set("X-Total-Count", strconv.FormatInt(total, 10))
	ctx.Response.Header.Set("X-Page", strconv.Itoa(pagination.Page))
	ctx.Response.Header.Set("X-Page-Size", strconv.Itoa(pagination.PageSize))

	// Marshal users to JSON with pretty printing
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal users: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleGetUser handles the GET /admin/users/{username} endpoint
func (h *AdminHandler) HandleGetUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get user from service
	user, err := h.userService.GetUser(ctx, tenant, realm, username)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user")
		return
	}

	if user == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Marshal user to JSON with pretty printing
	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal user: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleUpdateUser handles the PUT /admin/users/{username} endpoint
func (h *AdminHandler) HandleUpdateUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Parse request body
	var updateUser model.User
	if err := json.Unmarshal(ctx.PostBody(), &updateUser); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Update user through service
	user, err := h.userService.UpdateUser(ctx, tenant, realm, username, updateUser)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to update user")
		return
	}

	if user == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Marshal user to JSON with pretty printing
	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal user: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleDeleteUser handles the DELETE /admin/users/{username} endpoint
func (h *AdminHandler) HandleDeleteUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Delete user through service
	err := h.userService.DeleteUser(ctx, tenant, realm, username)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to delete user")
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}
