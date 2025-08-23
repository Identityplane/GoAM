package nodetotp

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func TestTOTPCreateAndVerifyFlow(t *testing.T) {
	// Arrange
	// Create a new user state with an example user
	testUser := &model.User{
		ID:          uuid.NewString(),
		Username:    "testuser",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Tenant:      "acme",
		Realm:       "customers",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create actual repository implementation using SQLite in-memory database
	sqliteDB, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	assert.NoError(t, err)
	defer sqliteDB.Close()

	// Run migrations to create tables
	err = sqlite_adapter.RunMigrations(sqliteDB)
	assert.NoError(t, err)

	userDB, err := sqlite_adapter.NewUserDB(sqliteDB)
	assert.NoError(t, err)

	userAttributeDB, err := sqlite_adapter.NewUserAttributeDB(sqliteDB)
	assert.NoError(t, err)

	userRepo := repository.NewUserRepository("acme", "customers", userDB, userAttributeDB)
	services := &model.Repositories{
		UserRepo: userRepo,
	}

	// Create test session with user in context
	session := &model.AuthenticationSession{
		User: testUser,
		Context: map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
		},
	}

	// Create TOTP create node
	createNode := &model.GraphNode{
		CustomConfig: map[string]string{
			"totpIssuer": "TestApp",
			"saveUser":   "true",
		},
	}

	// Act 1: Run CreateTOTP Node -> Expect TOTP prompt with secret and QR code
	result, err := RunTOTPCreateNode(session, createNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "string", result.Prompts["totpVerification"])

	// Verify that TOTP secret and image URL are in the context
	assert.NotEmpty(t, session.Context["totpSecret"])
	assert.NotEmpty(t, session.Context["totpImageUrl"])
	assert.NotEmpty(t, session.Context["totpIssuer"])

	// Store the generated TOTP secret for later verification
	totpSecret := session.Context["totpSecret"]

	// Act 2: Calculate a TOTP value using the secret and verify the prompt
	// Generate a valid TOTP code using the secret
	validTOTPCode, err := totp.GenerateCode(totpSecret, time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, validTOTPCode)

	// Now provide the valid TOTP code to complete the creation
	result, err = RunTOTPCreateNode(session, createNode, map[string]string{
		"totpVerification": validTOTPCode,
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)
	assert.Empty(t, result.Prompts)

	// Verify that the TOTP attribute was added to the user
	totpAttrs := testUser.GetAttributesByType(model.AttributeTypeTOTP)
	t.Logf("TOTP attributes found: %d", len(totpAttrs))
	t.Logf("User attributes: %+v", testUser.UserAttributes)
	assert.Len(t, totpAttrs, 1)

	// Get the TOTP attribute value
	totpValue, _, err := model.GetAttribute[model.TOTPAttributeValue](testUser, model.AttributeTypeTOTP)
	assert.NoError(t, err)
	assert.NotNil(t, totpValue)
	assert.Equal(t, totpSecret, totpValue.SecretKey)
	assert.False(t, totpValue.Locked)
	assert.Equal(t, 0, totpValue.FailedAttempts)

	// Act 3: Test TOTP verification with invalid code
	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			"max_failed_attempts": "5",
		},
	}

	// First, test with no input - should prompt for verification code
	result, err = RunTOTPVerifyNode(session, verifyNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "string", result.Prompts["totpVerification"])

	// Test with invalid TOTP code - should increment failed counter
	invalidCode := "000000"

	result, err = RunTOTPVerifyNode(session, verifyNode, map[string]string{
		"totpVerification": invalidCode,
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "failure", result.Condition)

	// Verify that failed attempts counter was incremented
	totpValue, _, err = model.GetAttribute[model.TOTPAttributeValue](testUser, model.AttributeTypeTOTP)
	assert.NoError(t, err)
	assert.NotNil(t, totpValue)
	t.Logf("TOTP value after first invalid attempt: %+v", totpValue)
	assert.Equal(t, 1, totpValue.FailedAttempts)
	assert.False(t, totpValue.Locked)

	// Test with another invalid code to increment counter further
	result, err = RunTOTPVerifyNode(session, verifyNode, map[string]string{
		"totpVerification": "111111",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "failure", result.Condition)

	// Verify counter increased again
	totpValue, _, err = model.GetAttribute[model.TOTPAttributeValue](testUser, model.AttributeTypeTOTP)
	assert.NoError(t, err)
	assert.NotNil(t, totpValue)
	assert.Equal(t, 2, totpValue.FailedAttempts)
	assert.False(t, totpValue.Locked)

	// Act 4: Test with valid TOTP code - should reset counter and return success
	validCode, err := totp.GenerateCode(totpSecret, time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, validCode)

	result, err = RunTOTPVerifyNode(session, verifyNode, map[string]string{
		"totpVerification": validCode,
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)

	// Verify that failed attempts counter was reset
	totpValue, _, err = model.GetAttribute[model.TOTPAttributeValue](testUser, model.AttributeTypeTOTP)
	assert.NoError(t, err)
	assert.NotNil(t, totpValue)
	assert.Equal(t, 0, totpValue.FailedAttempts)
	assert.False(t, totpValue.Locked)
}

func TestTOTPCreateNodeWithoutUser(t *testing.T) {
	// Test that TOTP create node fails when no user is in context
	session := &model.AuthenticationSession{
		Context: map[string]string{},
		User:    nil, // No user in context
	}

	createNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunTOTPCreateNode(session, createNode, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found in context")
}

func TestTOTPVerifyNodeWithoutUser(t *testing.T) {
	// Test that TOTP verify node fails when no user is in context
	session := &model.AuthenticationSession{
		Context: map[string]string{},
		User:    nil, // No user in context
	}

	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunTOTPVerifyNode(session, verifyNode, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found in context")
}

func TestTOTPVerifyNodeWithNoTOTPAttribute(t *testing.T) {
	// Test TOTP verify node when user has no TOTP attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunTOTPVerifyNode(session, verifyNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "no_totp", result.Condition)
}

func TestTOTPCreateNodeWithInvalidVerification(t *testing.T) {
	// Test TOTP create node with invalid verification code
	testUser := &model.User{
		ID:          uuid.NewString(),
		Username:    "testuser",
		DisplayName: "Test User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	session := &model.AuthenticationSession{
		User: testUser,
		Context: map[string]string{
			"username": "testuser",
		},
	}

	createNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	// First, get the TOTP secret
	result, err := RunTOTPCreateNode(session, createNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, session.Context["totpSecret"])

	// Now try with invalid verification code
	result, err = RunTOTPCreateNode(session, createNode, map[string]string{
		"totpVerification": "000000",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Condition)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "string", result.Prompts["totpVerification"])
}

func TestTOTPVerifyNodeMaxFailedAttempts(t *testing.T) {
	// Test that TOTP gets locked after max failed attempts
	testUser := &model.User{
		ID:        uuid.NewString(),
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add TOTP attribute to user with a proper Index
	totpAttribute := &model.UserAttribute{
		ID:    uuid.NewString(),
		Type:  model.AttributeTypeTOTP,
		Index: "TESTSECRET123", // Set the Index field
		Value: model.TOTPAttributeValue{
			SecretKey:      "TESTSECRET123",
			Locked:         false,
			FailedAttempts: 0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	testUser.AddAttribute(totpAttribute)

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			"max_failed_attempts": "3", // Set max to 3
		},
	}

	// Create actual repository implementation using SQLite in-memory database
	sqliteDB, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	assert.NoError(t, err)
	defer sqliteDB.Close()

	// Run migrations to create tables
	err = sqlite_adapter.RunMigrations(sqliteDB)
	assert.NoError(t, err)

	userDB, err := sqlite_adapter.NewUserDB(sqliteDB)
	assert.NoError(t, err)

	userAttributeDB, err := sqlite_adapter.NewUserAttributeDB(sqliteDB)
	assert.NoError(t, err)

	userRepo := repository.NewUserRepository("acme", "customers", userDB, userAttributeDB)
	services := &model.Repositories{
		UserRepo: userRepo,
	}

	// Try 3 invalid codes to reach max failed attempts
	for i := 0; i < 3; i++ {
		result, err := RunTOTPVerifyNode(session, verifyNode, map[string]string{
			"totpVerification": "000000",
		}, services)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "failure", result.Condition)
	}

	// Verify that TOTP is now locked
	totpValue, _, err := model.GetAttribute[model.TOTPAttributeValue](testUser, model.AttributeTypeTOTP)
	assert.NoError(t, err)
	assert.NotNil(t, totpValue)
	assert.Equal(t, 3, totpValue.FailedAttempts)
	assert.True(t, totpValue.Locked)
}
