package admin_api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"

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

// @Summary List users
// @Description Get a paginated list of users
// @Tags Users
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
// @Router /admin/{tenant}/{realm}/users [get]
func HandleListUsers(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
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
	users, total, err := service.GetServices().UserService.ListUsers(ctx, tenant, realm, service.PaginationParams{
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

	// If users is empty, return an empty array
	if len(users) == 0 {
		users = []model.User{}
	}

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
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Success 200 {object} model.User
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{username} [get]
func HandleGetUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get user from service
	user, err := service.GetServices().UserService.GetUser(ctx, tenant, realm, username)
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
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Param user body model.User true "User object"
// @Success 201 {object} model.User
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{username} [post]
func HandleCreateUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Parse request body
	var createUser model.User
	if err := (&createUser).UnmarshalJSON(ctx.PostBody()); err != nil {
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
	user, err := service.GetServices().UserService.CreateUser(ctx, tenant, realm, createUser)
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
// @Tags Users
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
// @Router /admin/{tenant}/{realm}/users/{username} [put]
func HandleUpdateUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
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
	user, err := service.GetServices().UserService.UpdateUser(ctx, tenant, realm, username, updateUser)
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
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param username path string true "Username"
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{username} [delete]
func HandleDeleteUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	username := ctx.UserValue("username").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Delete user through service
	err := service.GetServices().UserService.DeleteUser(ctx, tenant, realm, username)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to delete user: " + err.Error())
		return
	}

	ctx.SetStatusCode(http.StatusNoContent)
}

// @Summary Get user statistics
// @Description Get user statistics for the realm
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} model.UserStats
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/stats [get]
func HandleGetUserStats(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get stats from service
	stats, err := service.GetServices().UserService.GetUserStats(ctx, tenant, realm)
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
