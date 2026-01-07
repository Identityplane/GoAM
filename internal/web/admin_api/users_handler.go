package admin_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/internal/web/webutils"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"

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

type UserJson struct {
	model.User
	Url string `json:"url"`
}

type UserWithAttributesJson struct {
	UserJson
	UserAttributes []*model.UserAttribute `json:"user_attributes"`
}

type UserFlatJson struct {
	model.UserFlat
	Url string `json:"url"`
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
	users, total, err := service.GetServices().UserService.ListUsers(ctx, tenant, realm, services_interface.PaginationParams{
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

	// Load attributes for each user to populate UserFlat fields
	userFlatJsons := []UserFlatJson{}
	for _, user := range users {
		// Get user with attributes to ensure we have all data for UserFlat conversion
		userWithAttrs, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, user.ID)
		if err != nil {
			// If we can't load attributes, still include the user but without attribute data
			userFlatJsons = append(userFlatJsons, UserToUserFlatJson(ctx, &user))
			continue
		}
		if userWithAttrs != nil {
			userFlatJsons = append(userFlatJsons, UserToUserFlatJson(ctx, userWithAttrs))
		} else {
			// Fallback to user without attributes
			userFlatJsons = append(userFlatJsons, UserToUserFlatJson(ctx, &user))
		}
	}

	// Create response
	response := PagedResponse{
		Data: userFlatJsons,
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
// @Description Get a specific user by id
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Success 200 {object} model.UserFlat
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id} [get]
func HandleGetUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	id := ctx.UserValue("id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get user with attributes (needed for UserFlat conversion)
	user, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, id)
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

	// Convert to UserFlat and add URL
	result := UserToUserFlatJson(ctx, user)

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(result, "", "  ")
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
// @Param id path string true "User ID"
// @Param user body model.UserFlat true "User object"
// @Success 201 {object} model.UserFlat
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id} [post]
func HandleCreateUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	id := ctx.UserValue("id").(string)

	// Validate realm
	if !validateRealm(ctx, tenant, realm) {
		return
	}

	// Parse request body as UserFlat
	var createUserFlat model.UserFlat
	if err := (&createUserFlat).UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Take the id in the url if the id is not set in the body
	if createUserFlat.ID == "" {
		createUserFlat.ID = id
	}

	// Check if id matched the path parameter
	if createUserFlat.ID != id {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("ID does not match path parameter")
		return
	}

	// Convert UserFlat to User
	createUser := createUserFlat.ToUser()

	// Create user with attributes through service
	user, err := service.GetServices().UserService.CreateUserWithAttributes(ctx, tenant, realm, *createUser)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to create user: " + err.Error())
		return
	}

	// Get user with attributes to ensure all attributes are loaded
	userWithAttrs, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, user.ID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user: " + err.Error())
		return
	}

	// Write response
	writeUserFlatResponse(ctx, userWithAttrs, http.StatusCreated)
}

// @Summary Update user
// @Description Update an existing user (replaces all attributes)
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Param user body model.UserFlat true "User object"
// @Success 200 {object} model.UserFlat
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id} [put]
func HandleUpdateUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	id := ctx.UserValue("id").(string)

	// Validate realm
	if !validateRealm(ctx, tenant, realm) {
		return
	}

	// Get existing user with attributes to preserve attribute IDs
	existingUserWithAttrs, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, id)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user: " + err.Error())
		return
	}

	if existingUserWithAttrs == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Parse request body as UserFlat
	var updateUserFlat model.UserFlat
	if err := (&updateUserFlat).UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Ensure ID matches path parameter
	if updateUserFlat.ID != "" && updateUserFlat.ID != id {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("ID does not match path parameter")
		return
	}
	updateUserFlat.ID = id

	// Convert UserFlat to User (this creates new attributes)
	newUser := updateUserFlat.ToUser()

	// Match new attributes with existing ones by type and preserve IDs
	for _, newAttr := range newUser.UserAttributes {
		// Find existing attribute of the same type
		for _, existingAttr := range existingUserWithAttrs.UserAttributes {
			if existingAttr.Type == newAttr.Type {
				// Preserve the existing attribute ID
				newAttr.ID = existingAttr.ID
				break
			}
		}
	}

	// Update user with attributes through service
	user, err := service.GetServices().UserService.UpdateUserWithAttributes(ctx, tenant, realm, *newUser)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to update user: " + err.Error())
		return
	}

	// Get updated user with attributes to ensure we have the latest data
	userWithAttrs, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, user.ID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get updated user: " + err.Error())
		return
	}

	// Write response
	writeUserFlatResponse(ctx, userWithAttrs, http.StatusOK)
}

// @Summary Delete user
// @Description Delete a user
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id} [delete]
func HandleDeleteUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	id := ctx.UserValue("id").(string)

	// Lookup the loaded realm
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Delete user through service
	err := service.GetServices().UserService.DeleteUserByID(ctx, tenant, realm, id)
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

func UserToUserJson(ctx *fasthttp.RequestCtx, user *model.User) UserJson {

	adminBaseUrl := webutils.GetAdminBaseUrl(ctx)

	return UserJson{
		User: *user,
		Url:  fmt.Sprintf("%s/%s/%s/users/%s?include_attributes=true", adminBaseUrl, user.Tenant, user.Realm, user.ID),
	}
}

func UserToUserWithAttributesJson(ctx *fasthttp.RequestCtx, user *model.User) UserWithAttributesJson {
	return UserWithAttributesJson{
		UserJson:       UserToUserJson(ctx, user),
		UserAttributes: user.UserAttributes[:],
	}
}

func UserToUserFlatJson(ctx *fasthttp.RequestCtx, user *model.User) UserFlatJson {
	adminBaseUrl := webutils.GetAdminBaseUrl(ctx)
	userFlat := user.ToUserFlat()

	return UserFlatJson{
		UserFlat: *userFlat,
		Url:      fmt.Sprintf("%s/%s/%s/users/%s", adminBaseUrl, user.Tenant, user.Realm, user.ID),
	}
}

// writeUserFlatResponse writes a UserFlatJson response to the context
func writeUserFlatResponse(ctx *fasthttp.RequestCtx, user *model.User, statusCode int) {
	result := UserToUserFlatJson(ctx, user)

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(statusCode)
	ctx.SetBody(jsonData)
}

// validateRealm checks if the realm exists and returns an error response if not
func validateRealm(ctx *fasthttp.RequestCtx, tenant, realm string) bool {
	_, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return false
	}
	return true
}

// applyPatchToUser applies a patch UserFlat to an existing User, updating only the attributes that are present in the patch
// patchFields is a map of field names that were present in the JSON request
func applyPatchToUser(existingUser *model.User, patchFlat *model.UserFlat, patchFields map[string]bool) {
	// Update user status if provided
	if patchFields["status"] && patchFlat.Status != "" {
		existingUser.Status = patchFlat.Status
	}

	// Helper to find or create attribute by type
	findOrCreateAttribute := func(attrType string) *model.UserAttribute {
		for _, attr := range existingUser.UserAttributes {
			if attr.Type == attrType {
				return attr
			}
		}
		// Create new attribute if not found
		newAttr := &model.UserAttribute{
			ID:     "",
			Type:   attrType,
			Tenant: existingUser.Tenant,
			Realm:  existingUser.Realm,
			UserID: existingUser.ID,
		}
		existingUser.AddAttribute(newAttr)
		return newAttr
	}

	// Update email attribute if any email field is present
	hasEmailField := patchFields["email"] || patchFields["email_verified"] || patchFields["email_verified_at"]
	if hasEmailField {
		// Get existing email value or create new one
		emailValue, _, err := model.GetAttribute[model.EmailAttributeValue](existingUser, model.AttributeTypeEmail)
		if err != nil || emailValue == nil {
			emailValue = &model.EmailAttributeValue{}
		}

		// Update only the fields present in the patch
		if patchFields["email"] {
			emailValue.Email = patchFlat.Email
		}
		if patchFields["email_verified"] && patchFlat.EmailVerified != nil {
			emailValue.Verified = *patchFlat.EmailVerified
		}
		if patchFields["email_verified_at"] {
			emailValue.VerifiedAt = patchFlat.EmailVerifiedAt
		}

		// Find or create the attribute and set the value
		emailAttr := findOrCreateAttribute(model.AttributeTypeEmail)
		emailAttr.Value = *emailValue
	}

	// Update phone attribute if any phone field is present
	hasPhoneField := patchFields["phone"] || patchFields["phone_verified"] || patchFields["phone_verified_at"]
	if hasPhoneField {
		// Get existing phone value or create new one
		phoneValue, _, err := model.GetAttribute[model.PhoneAttributeValue](existingUser, model.AttributeTypePhone)
		if err != nil || phoneValue == nil {
			phoneValue = &model.PhoneAttributeValue{}
		}

		// Update only the fields present in the patch
		if patchFields["phone"] {
			phoneValue.Phone = patchFlat.Phone
		}
		if patchFields["phone_verified"] && patchFlat.PhoneVerified != nil {
			phoneValue.Verified = *patchFlat.PhoneVerified
		}
		if patchFields["phone_verified_at"] {
			phoneValue.VerifiedAt = patchFlat.PhoneVerifiedAt
		}

		// Find or create the attribute and set the value
		phoneAttr := findOrCreateAttribute(model.AttributeTypePhone)
		phoneAttr.Value = *phoneValue
	}

	// Update username attribute if any username field is present
	hasUsernameField := patchFields["preferred_username"] ||
		patchFields["website"] ||
		patchFields["zoneinfo"] ||
		patchFields["birthdate"] ||
		patchFields["gender"] ||
		patchFields["profile"] ||
		patchFields["given_name"] ||
		patchFields["middle_name"] ||
		patchFields["locale"] ||
		patchFields["picture"] ||
		patchFields["name"] ||
		patchFields["nickname"] ||
		patchFields["family_name"]

	if hasUsernameField {
		// Get existing username value or create new one
		usernameValue, _, err := model.GetAttribute[model.UsernameAttributeValue](existingUser, model.AttributeTypeUsername)
		if err != nil || usernameValue == nil {
			usernameValue = &model.UsernameAttributeValue{}
		}

		// Update only the fields present in the patch
		if patchFields["preferred_username"] {
			usernameValue.PreferredUsername = patchFlat.PreferredUsername
		}
		if patchFields["website"] {
			usernameValue.Website = patchFlat.Website
		}
		if patchFields["zoneinfo"] {
			usernameValue.Zoneinfo = patchFlat.Zoneinfo
		}
		if patchFields["birthdate"] {
			usernameValue.Birthdate = patchFlat.Birthdate
		}
		if patchFields["gender"] {
			usernameValue.Gender = patchFlat.Gender
		}
		if patchFields["profile"] {
			usernameValue.Profile = patchFlat.Profile
		}
		if patchFields["given_name"] {
			usernameValue.GivenName = patchFlat.GivenName
		}
		if patchFields["middle_name"] {
			usernameValue.MiddleName = patchFlat.MiddleName
		}
		if patchFields["locale"] {
			usernameValue.Locale = patchFlat.Locale
		}
		if patchFields["picture"] {
			usernameValue.Picture = patchFlat.Picture
		}
		if patchFields["name"] {
			usernameValue.Name = patchFlat.Name
		}
		if patchFields["nickname"] {
			usernameValue.Nickname = patchFlat.Nickname
		}
		if patchFields["family_name"] {
			usernameValue.FamilyName = patchFlat.FamilyName
		}

		// Find or create the attribute and set the value
		usernameAttr := findOrCreateAttribute(model.AttributeTypeUsername)
		usernameAttr.Value = *usernameValue
	}
}

// @Summary Patch user
// @Description Partially update an existing user (only updates attributes present in the request)
// @Tags Users
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param id path string true "User ID"
// @Param user body model.UserFlat true "User fields to update"
// @Success 200 {object} model.UserFlat
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/users/{id} [patch]
func HandlePatchUser(ctx *fasthttp.RequestCtx) {
	// Get path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	id := ctx.UserValue("id").(string)

	// Validate realm
	if !validateRealm(ctx, tenant, realm) {
		return
	}

	// Get existing user with attributes
	existingUser, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, id)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user: " + err.Error())
		return
	}

	if existingUser == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("User not found")
		return
	}

	// Parse request body to detect which fields are present
	var patchData map[string]interface{}
	if err := json.Unmarshal(ctx.PostBody(), &patchData); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Create a map to track which fields were present in the JSON
	patchFields := make(map[string]bool)
	for key := range patchData {
		patchFields[key] = true
	}

	// Parse request body as UserFlat patch
	var patchFlat model.UserFlat
	if err := (&patchFlat).UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBodyString("Invalid JSON: " + err.Error())
		return
	}

	// Apply patch to existing user
	applyPatchToUser(existingUser, &patchFlat, patchFields)

	// Update user with attributes
	updatedUser, err := service.GetServices().UserService.UpdateUserWithAttributes(ctx, tenant, realm, *existingUser)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to update user: " + err.Error())
		return
	}

	// Get updated user with attributes to ensure we have the latest data
	userWithAttrs, err := service.GetServices().UserService.GetUserWithAttributesByID(ctx, tenant, realm, updatedUser.ID)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get updated user: " + err.Error())
		return
	}

	// Write response
	writeUserFlatResponse(ctx, userWithAttrs, http.StatusOK)
}
