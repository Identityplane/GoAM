package admin_api

import (
	"encoding/json"
	"goiam/internal/db/model"
	"goiam/internal/db/service"
	"goiam/internal/realms"
	"net/http"
	"strconv"

	"github.com/valyala/fasthttp"
)

// PagedResponse represents a paginated API response
type PagedResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// Handler handles admin API endpoints
type Handler struct {
	userService service.UserAdminService
}

// New creates a new admin API handler
func New(userService service.UserAdminService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(100)
// @Success 200 {object} PagedResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/users [get]
func (h *Handler) HandleListUsers(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 100 // default page size

	if pageStr := string(ctx.QueryArgs().Peek("page")); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := string(ctx.QueryArgs().Peek("page_size")); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get users from service
	users, total, err := h.userService.ListUsers(ctx, tenant, realm, service.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to list users: " + err.Error())
		return
	}

	// Calculate total pages
	totalPages := (int(total) + pageSize - 1) / pageSize

	// Create response
	response := PagedResponse{
		Data: users,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// @Summary Get user
// @Description Get a specific user by username
// @Tags users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Success 200 {object} model.User
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/users/{username} [get]
func (h *Handler) HandleGetUser(ctx *fasthttp.RequestCtx) {
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
		ctx.SetBodyString("Failed to get user: " + err.Error())
		return
	}

	if user == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Param user body model.User true "User object"
// @Success 201 {object} model.User
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/users/{username} [post]
func (h *Handler) HandleCreateUser(ctx *fasthttp.RequestCtx) {
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
	var createUser model.User
	if err := json.Unmarshal(ctx.PostBody(), &createUser); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Check if username matched the path parameter
	if createUser.Username != username {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Username does not match path parameter")
		return
	}

	// Create user through service
	user, err := h.userService.CreateUser(ctx, tenant, realm, createUser)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to create user: " + err.Error())
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body with new user details
	ctx.SetBody(jsonData)
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusCreated)
}

// @Summary Update user
// @Description Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Param user body model.User true "User object"
// @Success 200 {object} model.User
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/users/{username} [put]
func (h *Handler) HandleUpdateUser(ctx *fasthttp.RequestCtx) {
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
		ctx.SetBodyString("Failed to update user: " + err.Error())
		return
	}

	if user == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// @Summary Delete user
// @Description Delete a user
// @Tags users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/users/{username} [delete]
func (h *Handler) HandleDeleteUser(ctx *fasthttp.RequestCtx) {
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
		ctx.SetBodyString("Failed to delete user: " + err.Error())
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}

// @Summary Get user statistics
// @Description Get user statistics for the realm
// @Tags users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} model.UserStats
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/users/stats [get]
func (h *Handler) HandleGetUserStats(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get stats from service
	stats, err := h.userService.GetUserStats(ctx, tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user stats: " + err.Error())
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
