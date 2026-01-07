package node_yubico

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/db/sqlite_adapter"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockYubicoAPI implements the yubicoApiInterface for testing
type MockYubicoAPI struct {
	shouldSucceed bool
	shouldError   bool
	errorMessage  string
	publicId      string
}

func (m *MockYubicoAPI) Verify(id, otp, nonce string) (YubicoApiResponse, error) {
	if m.shouldError {
		return YubicoApiResponse{}, assert.AnError
	}
	if m.shouldSucceed {
		return YubicoApiResponse{
			OTP:    otp,
			Nonce:  nonce,
			Status: "OK",
			Sl:     100,
		}, nil
	}
	// Return a response indicating validation failure
	return YubicoApiResponse{
		OTP:    otp,
		Nonce:  nonce,
		Status: "BAD_OTP",
		Sl:     100,
	}, nil
}

func TestYubicoCreateHasVerifyFlow(t *testing.T) {
	// Arrange
	// Create a new user state with an example user
	testUser := &model.User{
		ID:        uuid.NewString(),
		Tenant:    "acme",
		Realm:     "customers",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

	// Create a new empty user in the database
	userRepo.Create(context.Background(), testUser)

	// Mock the Yubico verifier
	mockAPI := &MockYubicoAPI{
		shouldSucceed: true,
		shouldError:   false,
		publicId:      "cccccckdvvul",
	}

	// Override the getYubikeyVerifier function for testing
	originalGetYubikeyVerifier := getYubikeyVerifier
	getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
		return &YubicoVerifier{
			apiUrl:    apiUrl,
			clientId:  clientId,
			apiKey:    apiKey,
			yubicoApi: mockAPI,
		}
	}
	defer func() {
		// Restore original function
		getYubikeyVerifier = originalGetYubikeyVerifier
	}()

	// Create Yubico create node
	createNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_SKIP_SAVE_USER:      "false",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	// Act 1: Run Create Yubikey Node -> Expect Yubikey OTP prompt
	result, err := RunYubicoCreateNode(session, createNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "string", result.Prompts["yubikeyOtpVerification"])

	// Act 2: Provide valid OTP to complete the creation
	mockAPI.shouldSucceed = true
	mockAPI.publicId = "cccccckdvvul"

	result, err = RunYubicoCreateNode(session, createNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateSuccess, result.Condition)
	assert.Empty(t, result.Prompts)

	// Verify that the Yubikey attribute was added to the user
	yubikeyAttrs := testUser.GetAttributesByType(model.AttributeTypeYubico)
	assert.Len(t, yubikeyAttrs, 1)

	// Get the Yubikey attribute value
	yubikeyValue, _, err := model.GetAttribute[model.YubicoAttributeValue](testUser, model.AttributeTypeYubico)
	assert.NoError(t, err)
	assert.NotNil(t, yubikeyValue)
	assert.Equal(t, "cccccckdvvul", yubikeyValue.PublicID)
	assert.False(t, yubikeyValue.Locked)
	assert.Equal(t, 0, yubikeyValue.FailedAttempts)

	// Act 3: Run Has Yubikey Node -> Should return Yes
	hasNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	result, err = RunHasYubicoNode(session, hasNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateYes, result.Condition)

	// Act 4: Run Yubikey Verify Node with invalid OTP
	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	// First, test with no input - should prompt for verification code
	result, err = RunYubicoVerifyNode(session, verifyNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "string", result.Prompts["yubikeyOtpVerification"])

	// Test with invalid OTP - should return failure
	mockAPI.shouldSucceed = false
	mockAPI.shouldError = false

	result, err = RunYubicoVerifyNode(session, verifyNode, map[string]string{
		"yubikeyOtpVerification": "invalidotp",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateFailure, result.Condition)

	// Act 5: Run Yubikey Verify Node with valid OTP
	mockAPI.shouldSucceed = true
	mockAPI.publicId = "cccccckdvvul"

	result, err = RunYubicoVerifyNode(session, verifyNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateSuccess, result.Condition)
}

func TestYubicoCreateNodeWithoutUser(t *testing.T) {
	// Test that Yubico create node fails when no user is in context
	session := &model.AuthenticationSession{
		Context: map[string]string{},
		User:    nil, // No user in context
	}

	createNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:   "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID: "12345",
			CONFIG_YUBICO_API_KEY:   "dGVzdC1hcGkta2V5",
		},
	}

	services := &model.Repositories{}

	result, err := RunYubicoCreateNode(session, createNode, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user repository is not set")
}

func TestYubicoCreateNodeMissingConfig(t *testing.T) {
	// Test that Yubico create node fails when required config is missing
	testUser := &model.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	// Missing client ID
	createNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL: "https://api.yubico.com/wsapi/2.0/verify",
			// CONFIG_YUBICO_CLIENT_ID missing
			CONFIG_YUBICO_API_KEY: "dGVzdC1hcGkta2V5",
		},
	}

	services := &model.Repositories{}

	result, err := RunYubicoCreateNode(session, createNode, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "yubikey client id is required")
}

func TestYubicoHasNodeWithoutUser(t *testing.T) {
	// Test that Yubico has node fails when no user is in context
	session := &model.AuthenticationSession{
		Context: map[string]string{},
		User:    nil, // No user in context
	}

	hasNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunHasYubicoNode(session, hasNode, map[string]string{}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user repository is not set")
}

func TestYubicoHasNodeWithNoYubikey(t *testing.T) {
	// Test Yubico has node when user has no Yubikey attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	hasNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	services := &model.Repositories{}

	result, err := RunHasYubicoNode(session, hasNode, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateNo, result.Condition)
}

func TestYubicoVerifyNodeWithNoYubikey(t *testing.T) {
	// Test Yubico verify node when user has no Yubikey attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	services := &model.Repositories{}

	// Mock API for this test
	mockAPI := &MockYubicoAPI{
		shouldSucceed: true,
		shouldError:   false,
		publicId:      "cccccckdvvul",
	}

	originalGetYubikeyVerifier := getYubikeyVerifier
	getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
		return &YubicoVerifier{
			apiUrl:    apiUrl,
			clientId:  clientId,
			apiKey:    apiKey,
			yubicoApi: mockAPI,
		}
	}
	defer func() {
		getYubikeyVerifier = originalGetYubikeyVerifier
	}()

	result, err := RunYubicoVerifyNode(session, verifyNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateFailure, result.Condition)
}

func TestYubicoVerifyNodeWithInternalError(t *testing.T) {
	// Test Yubico verify node when internal error occurs
	testUser := &model.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add Yubikey attribute to user
	yubikeyAttribute := &model.UserAttribute{
		ID:     uuid.NewString(),
		UserID: testUser.ID,
		Type:   model.AttributeTypeYubico,
		Value: model.YubicoAttributeValue{
			PublicID:       "cccccckdvvul",
			Locked:         false,
			FailedAttempts: 0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	testUser.AddAttribute(yubikeyAttribute)

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	services := &model.Repositories{}

	// Mock API that returns internal error
	mockAPI := &MockYubicoAPI{
		shouldSucceed: false,
		shouldError:   true,
	}

	originalGetYubikeyVerifier := getYubikeyVerifier
	getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
		return &YubicoVerifier{
			apiUrl:    apiUrl,
			clientId:  clientId,
			apiKey:    apiKey,
			yubicoApi: mockAPI,
		}
	}
	defer func() {
		getYubikeyVerifier = originalGetYubikeyVerifier
	}()

	result, err := RunYubicoVerifyNode(session, verifyNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestYubicoCreateNodeWithUniqueCheck(t *testing.T) {
	// Test Yubico create node with unique check enabled
	testUser := &model.User{
		ID:        uuid.NewString(),
		Tenant:    "acme",
		Realm:     "customers",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

	// Create a new empty user in the database
	userRepo.Create(context.Background(), testUser)

	// Mock the Yubico verifier
	mockAPI := &MockYubicoAPI{
		shouldSucceed: true,
		shouldError:   false,
		publicId:      "cccccckdvvul",
	}

	// Override the getYubikeyVerifier function for testing
	originalGetYubikeyVerifier := getYubikeyVerifier
	getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
		return &YubicoVerifier{
			apiUrl:    apiUrl,
			clientId:  clientId,
			apiKey:    apiKey,
			yubicoApi: mockAPI,
		}
	}
	defer func() {
		// Restore original function
		getYubikeyVerifier = originalGetYubikeyVerifier
	}()

	// Create Yubico create node with unique check enabled
	createNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_SKIP_SAVE_USER:      "false",
			CONFIG_CHECK_YUBICO_UNIQUE: "true",
		},
	}

	// First, create a Yubikey successfully
	result, err := RunYubicoCreateNode(session, createNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateSuccess, result.Condition)

	// Now try to create another user with the same Yubikey
	anotherUser := &model.User{
		ID:        uuid.NewString(),
		Tenant:    "acme",
		Realm:     "customers",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userRepo.Create(context.Background(), anotherUser)

	anotherSession := &model.AuthenticationSession{
		User: anotherUser,
		Context: map[string]string{
			"username": "anotheruser",
			"email":    "another@example.com",
		},
	}

	// Try to create the same Yubikey for another user - should return existing
	result, err = RunYubicoCreateNode(anotherSession, createNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateExisting, result.Condition)
}

func TestYubicoVerifyNodeWithLockedYubikey(t *testing.T) {
	// Test Yubico verify node when user has a locked Yubikey
	testUser := &model.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add locked Yubikey attribute to user
	yubikeyAttribute := &model.UserAttribute{
		ID:     uuid.NewString(),
		UserID: testUser.ID,
		Type:   model.AttributeTypeYubico,
		Value: model.YubicoAttributeValue{
			PublicID:       "cccccckdvvul",
			Locked:         true, // Yubikey is locked
			FailedAttempts: 5,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	testUser.AddAttribute(yubikeyAttribute)

	session := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
	}

	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	services := &model.Repositories{}

	// Mock API for this test
	mockAPI := &MockYubicoAPI{
		shouldSucceed: true,
		shouldError:   false,
		publicId:      "cccccckdvvul",
	}

	originalGetYubikeyVerifier := getYubikeyVerifier
	getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
		return &YubicoVerifier{
			apiUrl:    apiUrl,
			clientId:  clientId,
			apiKey:    apiKey,
			yubicoApi: mockAPI,
		}
	}
	defer func() {
		getYubikeyVerifier = originalGetYubikeyVerifier
	}()

	// Test with valid OTP but locked Yubikey - should return locked
	result, err := RunYubicoVerifyNode(session, verifyNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateLocked, result.Condition)
}

func TestYubicoCreateVerifyFlowWithLockedYubikey(t *testing.T) {
	// Test complete flow: create Yubikey, lock it, then try to verify
	testUser := &model.User{
		ID:        uuid.NewString(),
		Tenant:    "acme",
		Realm:     "customers",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

	// Create a new empty user in the database
	userRepo.Create(context.Background(), testUser)

	// Mock the Yubico API
	mockAPI := &MockYubicoAPI{
		shouldSucceed: true,
		shouldError:   false,
		publicId:      "cccccckdvvul",
	}

	// Override the getYubikeyVerifier function for testing
	originalGetYubikeyVerifier := getYubikeyVerifier
	getYubikeyVerifier = func(apiUrl, clientId, apiKey string) *YubicoVerifier {
		return &YubicoVerifier{
			apiUrl:    apiUrl,
			clientId:  clientId,
			apiKey:    apiKey,
			yubicoApi: mockAPI,
		}
	}
	defer func() {
		// Restore original function
		getYubikeyVerifier = originalGetYubikeyVerifier
	}()

	// Create Yubico create node
	createNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_SKIP_SAVE_USER:      "false",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	// Step 1: Create a Yubikey successfully
	result, err := RunYubicoCreateNode(session, createNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateSuccess, result.Condition)

	// Step 2: Manually lock the Yubikey attribute (simulating admin action or failed attempts)
	yubikeyAttrs := testUser.GetAttributesByType(model.AttributeTypeYubico)
	assert.Len(t, yubikeyAttrs, 1)

	// Lock the Yubikey by directly modifying the attribute value
	for i := range testUser.UserAttributes {
		if testUser.UserAttributes[i].Type == model.AttributeTypeYubico {
			// Get the current value
			yubikeyValue, ok := testUser.UserAttributes[i].Value.(model.YubicoAttributeValue)
			assert.True(t, ok, "Expected YubicoAttributeValue")

			// Modify the value
			yubikeyValue.Locked = true
			yubikeyValue.FailedAttempts = 5

			// Update the attribute value
			testUser.UserAttributes[i].Value = yubikeyValue
			break
		}
	}

	// Save the updated user to the database
	err = services.UserRepo.CreateOrUpdate(context.Background(), testUser)
	assert.NoError(t, err)

	// Reload the user from the database to ensure we have the latest state
	reloadedUser, err := services.UserRepo.GetByID(context.Background(), testUser.ID)
	assert.NoError(t, err)
	session.User = reloadedUser

	// Verify that the reloaded user has the locked Yubikey
	yubikeyAttrs = reloadedUser.GetAttributesByType(model.AttributeTypeYubico)
	assert.Len(t, yubikeyAttrs, 1)
	yubikeyValue, _, err := model.GetAttribute[model.YubicoAttributeValue](reloadedUser, model.AttributeTypeYubico)
	assert.NoError(t, err)
	assert.True(t, yubikeyValue.Locked, "Yubikey should be locked after reload")

	// Step 3: Try to verify with the locked Yubikey
	verifyNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_YUBICO_API_URL:      "https://api.yubico.com/wsapi/2.0/verify",
			CONFIG_YUBICO_CLIENT_ID:    "12345",
			CONFIG_YUBICO_API_KEY:      "dGVzdC1hcGkta2V5",
			CONFIG_CHECK_YUBICO_UNIQUE: "false",
		},
	}

	result, err = RunYubicoVerifyNode(session, verifyNode, map[string]string{
		"yubikeyOtpVerification": "cccccckdvvulgjvtkjdhtlrbjjctggdihuevikehtlil",
	}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.ResultStateLocked, result.Condition)
}
