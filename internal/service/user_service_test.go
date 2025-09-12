package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestUserService creates a user service with in-memory SQLite database
func setupTestUserService(t *testing.T) (services_interface.UserAdminService, func()) {
	// Create SQLite in-memory database
	sqliteDB, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	require.NoError(t, err)

	// Run migrations to create tables
	err = sqlite_adapter.RunMigrations(sqliteDB)
	require.NoError(t, err)

	// Create database adapters
	userDB, err := sqlite_adapter.NewUserDB(sqliteDB)
	require.NoError(t, err)

	userAttributeDB, err := sqlite_adapter.NewUserAttributeDB(sqliteDB)
	require.NoError(t, err)

	// Create user service
	userService := NewUserService(userDB, userAttributeDB)

	// Return cleanup function
	cleanup := func() {
		sqliteDB.Close()
	}

	return userService, cleanup
}

func TestCRUD_CreateAndUpdateUserWithAttributes(t *testing.T) {
	ctx := context.Background()
	tenant := "test-tenant"
	realm := "test-realm"

	// Test cases
	testCases := []struct {
		name          string
		method        string
		user          model.User
		preCreateUser *model.User // User to create before the test (for update tests)
		expectedUser  *model.User // Expected user after operation (nil if expecting error or non-existing)
		expectError   bool
		expectSuccess bool
		description   string
	}{
		{
			name:          "CreateUserWithAttributes_NewUser_Success",
			method:        "CreateUserWithAttributes",
			user:          *createTestUser("", "newuser@example.com", "John", "Doe"),
			expectedUser:  createTestUser("", "newuser@example.com", "John", "Doe"),
			expectError:   false,
			expectSuccess: true,
			description:   "Create new user with attributes should succeed",
		},
		{
			name:          "CreateUserWithAttributes_ExistingUser_Error",
			method:        "CreateUserWithAttributes",
			user:          *createTestUser("existing-user-id", "existing@example.com", "Jane", "Smith"),
			preCreateUser: createTestUser("existing-user-id", "existing@example.com", "Jane", "Smith"),
			expectedUser:  createTestUser("existing-user-id", "existing@example.com", "Jane", "Smith"), // Should remain unchanged
			expectError:   true,
			expectSuccess: false,
			description:   "Create user with existing ID should fail",
		},
		{
			name:          "UpdateUserWithAttributes_NonExistingUser_Error",
			method:        "UpdateUserWithAttributes",
			user:          *createTestUser("non-existing-id", "update@example.com", "Update", "User"),
			expectedUser:  nil, // Should not exist
			expectError:   true,
			expectSuccess: false,
			description:   "Update non-existing user should fail",
		},
		{
			name:          "UpdateUserWithAttributes_ExistingUser_Success",
			method:        "UpdateUserWithAttributes",
			user:          *createTestUser("update-user-id", "updated@example.com", "Updated", "User"),
			preCreateUser: createTestUser("update-user-id", "original@example.com", "Original", "User"),
			expectedUser:  createTestUser("update-user-id", "updated@example.com", "Updated", "User"), // Should be updated
			expectError:   false,
			expectSuccess: true,
			description:   "Update existing user should succeed",
		},
		{
			name:          "CreateOrUpdateUserWithAttributes_NewUser_Success",
			method:        "CreateOrUpdateUserWithAttributes",
			user:          *createTestUser("", "upsert-new@example.com", "Upsert", "New"),
			expectedUser:  createTestUser("", "upsert-new@example.com", "Upsert", "New"),
			expectError:   false,
			expectSuccess: true,
			description:   "CreateOrUpdate new user should succeed",
		},
		{
			name:          "CreateOrUpdateUserWithAttributes_ExistingUser_Success",
			method:        "CreateOrUpdateUserWithAttributes",
			user:          *createTestUser("upsert-existing-id", "upsert-updated@example.com", "Upsert", "Updated"),
			preCreateUser: createTestUser("upsert-existing-id", "upsert-original@example.com", "Upsert", "Original"),
			expectedUser:  createTestUser("upsert-existing-id", "upsert-updated@example.com", "Upsert", "Updated"), // Should be updated
			expectError:   false,
			expectSuccess: true,
			description:   "CreateOrUpdate existing user should succeed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh user service for each test case
			userService, cleanup := setupTestUserService(t)
			defer cleanup()

			// Pre-create user if needed
			if tc.preCreateUser != nil {
				_, err := userService.CreateUserWithAttributes(ctx, tenant, realm, *tc.preCreateUser)
				require.NoError(t, err, "Failed to pre-create user for test case: %s", tc.name)

				// For update tests, get the actual user from database to use correct attribute IDs
				if tc.method == "UpdateUserWithAttributes" || tc.method == "CreateOrUpdateUserWithAttributes" {
					actualPreCreatedUser, err := userService.GetUserWithAttributesByID(ctx, tenant, realm, tc.preCreateUser.ID)
					require.NoError(t, err, "Failed to retrieve pre-created user for test case: %s", tc.name)
					require.NotNil(t, actualPreCreatedUser, "Pre-created user should exist for test case: %s", tc.name)

					// Update the test user to use the actual attribute IDs from the database
					for i, attr := range tc.user.UserAttributes {
						// Find the corresponding attribute in the pre-created user by type
						for _, preCreatedAttr := range actualPreCreatedUser.UserAttributes {
							if preCreatedAttr.Type == attr.Type {
								tc.user.UserAttributes[i].ID = preCreatedAttr.ID
								break
							}
						}
					}
				}
			}

			// Execute the test method
			var result *model.User
			var err error

			switch tc.method {
			case "CreateUserWithAttributes":
				result, err = userService.CreateUserWithAttributes(ctx, tenant, realm, tc.user)
			case "UpdateUserWithAttributes":
				result, err = userService.UpdateUserWithAttributes(ctx, tenant, realm, tc.user)
			case "CreateOrUpdateUserWithAttributes":
				result, err = userService.CreateOrUpdateUserWithAttributes(ctx, tenant, realm, tc.user)
			default:
				t.Fatalf("Unknown method: %s", tc.method)
			}

			// Verify expectations
			if tc.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tc.name)
				assert.Nil(t, result, "Expected nil result when error occurs for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for test case: %s", tc.name)
				assert.NotNil(t, result, "Expected non-nil result for test case: %s", tc.name)

				if tc.expectSuccess {
					// Verify user properties
					assert.Equal(t, tenant, result.Tenant, "Tenant should match for test case: %s", tc.name)
					assert.Equal(t, realm, result.Realm, "Realm should match for test case: %s", tc.name)
					assert.NotEmpty(t, result.ID, "User ID should not be empty for test case: %s", tc.name)
					assert.Equal(t, "active", result.Status, "User status should be active for test case: %s", tc.name)

					// Verify attributes were created/updated
					assert.NotEmpty(t, result.UserAttributes, "User should have attributes for test case: %s", tc.name)

					// Verify specific attributes
					emailAttr := findAttributeByType(result.UserAttributes, "email")
					assert.NotNil(t, emailAttr, "Email attribute should exist for test case: %s", tc.name)
					assert.Equal(t, tc.user.UserAttributes[0].Value, emailAttr.Value, "Email value should match for test case: %s", tc.name)

					firstNameAttr := findAttributeByType(result.UserAttributes, "first_name")
					assert.NotNil(t, firstNameAttr, "First name attribute should exist for test case: %s", tc.name)
					assert.Equal(t, tc.user.UserAttributes[1].Value, firstNameAttr.Value, "First name value should match for test case: %s", tc.name)

					lastNameAttr := findAttributeByType(result.UserAttributes, "last_name")
					assert.NotNil(t, lastNameAttr, "Last name attribute should exist for test case: %s", tc.name)
					assert.Equal(t, tc.user.UserAttributes[2].Value, lastNameAttr.Value, "Last name value should match for test case: %s", tc.name)
				}
			}

			// Verify database state by querying the user
			if tc.expectedUser != nil {
				// Determine which user ID to query - use result ID if available, otherwise expectedUser ID
				var userIDToQuery string
				if result != nil && result.ID != "" {
					userIDToQuery = result.ID
				} else {
					userIDToQuery = tc.expectedUser.ID
				}

				// We expect a user to exist, verify it matches expectedUser
				actualUser, err := userService.GetUserWithAttributesByID(ctx, tenant, realm, userIDToQuery)
				require.NoError(t, err, "Failed to retrieve user from database for test case: %s", tc.name)
				assert.NotNil(t, actualUser, "Expected user should exist in database for test case: %s", tc.name)

				// Verify user properties match expected
				if result != nil && result.ID != "" {
					assert.Equal(t, result.ID, actualUser.ID, "User ID should match result for test case: %s", tc.name)
				} else {
					assert.Equal(t, tc.expectedUser.ID, actualUser.ID, "User ID should match expected for test case: %s", tc.name)
				}
				assert.Equal(t, tenant, actualUser.Tenant, "User tenant should match for test case: %s", tc.name)
				assert.Equal(t, realm, actualUser.Realm, "User realm should match for test case: %s", tc.name)
				assert.Equal(t, "active", actualUser.Status, "User status should be active for test case: %s", tc.name)

				// Verify attributes match expected
				assert.NotEmpty(t, actualUser.UserAttributes, "User should have attributes in database for test case: %s", tc.name)

				// Verify specific attributes - use result user data if available, otherwise expected user data
				var expectedUserData *model.User
				if result != nil && len(result.UserAttributes) > 0 {
					expectedUserData = result
				} else {
					expectedUserData = tc.expectedUser
				}

				expectedEmailAttr := findAttributeByType(expectedUserData.UserAttributes, "email")
				actualEmailAttr := findAttributeByType(actualUser.UserAttributes, "email")
				assert.NotNil(t, expectedEmailAttr, "Expected user should have email attribute for test case: %s", tc.name)
				assert.NotNil(t, actualEmailAttr, "Actual user should have email attribute for test case: %s", tc.name)
				if expectedEmailAttr != nil && actualEmailAttr != nil {
					assert.Equal(t, expectedEmailAttr.Value, actualEmailAttr.Value, "Email attribute should match expected for test case: %s", tc.name)
				}

				expectedFirstNameAttr := findAttributeByType(expectedUserData.UserAttributes, "first_name")
				actualFirstNameAttr := findAttributeByType(actualUser.UserAttributes, "first_name")
				assert.NotNil(t, expectedFirstNameAttr, "Expected user should have first_name attribute for test case: %s", tc.name)
				assert.NotNil(t, actualFirstNameAttr, "Actual user should have first_name attribute for test case: %s", tc.name)
				if expectedFirstNameAttr != nil && actualFirstNameAttr != nil {
					assert.Equal(t, expectedFirstNameAttr.Value, actualFirstNameAttr.Value, "First name attribute should match expected for test case: %s", tc.name)
				}

				expectedLastNameAttr := findAttributeByType(expectedUserData.UserAttributes, "last_name")
				actualLastNameAttr := findAttributeByType(actualUser.UserAttributes, "last_name")
				assert.NotNil(t, expectedLastNameAttr, "Expected user should have last_name attribute for test case: %s", tc.name)
				assert.NotNil(t, actualLastNameAttr, "Actual user should have last_name attribute for test case: %s", tc.name)
				if expectedLastNameAttr != nil && actualLastNameAttr != nil {
					assert.Equal(t, expectedLastNameAttr.Value, actualLastNameAttr.Value, "Last name attribute should match expected for test case: %s", tc.name)
				}
			} else {
				// We expect no user to exist (for error cases or non-existing user tests)
				// Try to get the user that was attempted to be created/updated
				var userIDToCheck string
				if tc.user.ID != "" {
					userIDToCheck = tc.user.ID
				} else if tc.preCreateUser != nil {
					userIDToCheck = tc.preCreateUser.ID
				}

				if userIDToCheck != "" {
					actualUser, err := userService.GetUserWithAttributesByID(ctx, tenant, realm, userIDToCheck)
					if tc.expectError {
						// For error cases, user should either not exist or remain unchanged
						if err != nil {
							// User doesn't exist, which is expected for some error cases
							assert.Nil(t, actualUser, "User should not exist for error test case: %s", tc.name)
						} else {
							// User exists, verify it matches the pre-created user (unchanged)
							assert.NotNil(t, actualUser, "User should exist but be unchanged for error test case: %s", tc.name)
							if tc.preCreateUser != nil {
								assert.Equal(t, tc.preCreateUser.ID, actualUser.ID, "User ID should match pre-created user for error test case: %s", tc.name)
								// Verify attributes match pre-created user
								expectedEmailAttr := findAttributeByType(tc.preCreateUser.UserAttributes, "email")
								actualEmailAttr := findAttributeByType(actualUser.UserAttributes, "email")
								if expectedEmailAttr != nil && actualEmailAttr != nil {
									assert.Equal(t, expectedEmailAttr.Value, actualEmailAttr.Value, "Email should remain unchanged for error test case: %s", tc.name)
								}
							}
						}
					}
				}
			}
		})
	}
}

// Helper function to create a test user with attributes
func createTestUser(id, email, firstName, lastName string) *model.User {
	userID := id
	if userID == "" {
		userID = uuid.NewString()
	}

	now := time.Now()
	user := &model.User{
		ID:        userID,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
		UserAttributes: []model.UserAttribute{
			{
				ID:    uuid.NewString(),
				Type:  "email",
				Value: email,
				Index: &email,
			},
			{
				ID:    uuid.NewString(),
				Type:  "first_name",
				Value: firstName,
			},
			{
				ID:    uuid.NewString(),
				Type:  "last_name",
				Value: lastName,
			},
		},
	}

	// Set tenant and realm for attributes
	for i := range user.UserAttributes {
		user.UserAttributes[i].UserID = user.ID
		user.UserAttributes[i].Tenant = "test-tenant"
		user.UserAttributes[i].Realm = "test-realm"
		user.UserAttributes[i].CreatedAt = now
		user.UserAttributes[i].UpdatedAt = now
	}

	return user
}

// Helper function to find attribute by type
func findAttributeByType(attributes []model.UserAttribute, attrType string) *model.UserAttribute {
	for i := range attributes {
		if attributes[i].Type == attrType {
			return &attributes[i]
		}
	}
	return nil
}
