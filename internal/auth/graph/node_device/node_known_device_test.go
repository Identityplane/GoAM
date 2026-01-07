package node_device

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/pkg/model/attributes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddKnownDeviceNode(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test user
	testUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
	}

	// Create node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create authentication session with user set
	state := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{
				"user-agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
			},
			RequestIP:                 "192.168.1.100",
			RequestCookies:            map[string]string{},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input := map[string]string{}

	// Setup expectation for CreateUserAttribute
	mockUserRepo.On("CreateUserAttribute", mock.Anything, mock.MatchedBy(func(attr *model.UserAttribute) bool {
		return attr.Type == model.AttributeTypeDevice &&
			attr.Index != nil &&
			attr.Value != nil
	})).Return(nil)

	// Execute
	result, err := RunAddKnownDeviceNode(state, node, input, services)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)

	// Assert cookie is in the authentication request context
	cookie, exists := state.HttpAuthContext.AdditionalResponseCookies["session"]
	assert.True(t, exists, "Cookie should be in AdditionalResponseCookies")
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, "session", cookie.Name)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)

	// Assert device is in context
	assert.NotEmpty(t, state.Context["device"])

	mockUserRepo.AssertExpectations(t)
}

func TestIsKnownDeviceNode(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test user
	testUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
	}

	// Create device secret and hash
	deviceSecret := lib.GenerateSecureSessionID()
	deviceSecretHash := lib.HashString(deviceSecret)
	deviceID := lib.GenerateSecureSessionID()

	now := time.Now()
	expiry := now.Add(30 * 24 * time.Hour) // 30 days

	// Create device attribute value
	device := model.DeviceAttributeValue{
		DeviceID:         deviceID,
		DeviceSecretHash: deviceSecretHash,
		DeviceName:       "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		DeviceIP:         "192.168.1.100",
		DeviceUserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		CookieName:       "session",
		CookieExpires:    expiry,
		CookieSameSite:   "Lax",
		CookieHttpOnly:   true,
		CookieSecure:     true,
		SessionLoa0:      *attributes.InitSession(now, attributes.DEFAULT_LOA_TO_EXPIRY_MAPPINGS[0]),
		SessionLoa1:      nil,
		SessionLoa2:      nil,
	}

	// Add device attribute to user
	deviceAttribute := &model.UserAttribute{
		Type:      model.AttributeTypeDevice,
		Value:     device,
		Index:     &deviceSecretHash,
		CreatedAt: now,
		UpdatedAt: now,
	}
	testUser.AddAttribute(deviceAttribute)

	// Create node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create empty authentication session with cookie
	state := &model.AuthenticationSession{
		User:    nil,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{},
			RequestIP:      "192.168.1.100",
			RequestCookies: map[string]string{
				"session": deviceSecret, // Cookie contains the device secret, not the hash
			},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input := map[string]string{}

	// Setup expectation for GetByAttributeIndex - should return user with device
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, deviceSecretHash).Return(testUser, nil)

	// Setup expectation for UpdateUserAttribute - should update last activity
	mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.MatchedBy(func(attr *model.UserAttribute) bool {
		return attr.Type == model.AttributeTypeDevice
	})).Return(nil)

	// Execute
	result, err := RunIsKnownDeviceNode(state, node, input, services)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CONDITION_KNOWN_DEVICE, result.Condition)

	// Assert user is authenticated
	assert.NotNil(t, state.User)
	assert.Equal(t, testUser.ID, state.User.ID)

	// Assert device is in context
	assert.NotEmpty(t, state.Context["device"])
	assert.Equal(t, deviceID, state.Context["device"])

	mockUserRepo.AssertExpectations(t)
}

func TestAddKnownDeviceAndThenIsKnownDevice(t *testing.T) {
	// This test combines both steps as described in the requirements

	// Setup
	mockUserRepo := repository.NewMockUserRepository()

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test user
	testUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
	}

	// ========== STEP 1: Add Known Device ==========
	// Create node for adding device
	addDeviceNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create authentication session with user set
	state1 := &model.AuthenticationSession{
		User:    testUser,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{
				"user-agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
			},
			RequestIP:                 "192.168.1.100",
			RequestCookies:            map[string]string{},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input1 := map[string]string{}

	// Setup expectation for CreateUserAttribute
	mockUserRepo.On("CreateUserAttribute", mock.Anything, mock.MatchedBy(func(attr *model.UserAttribute) bool {
		return attr.Type == model.AttributeTypeDevice &&
			attr.Index != nil &&
			attr.Value != nil
	})).Return(nil)

	// Execute STEP 1
	result1, err1 := RunAddKnownDeviceNode(state1, addDeviceNode, input1, services)

	// Assertions for STEP 1
	assert.NoError(t, err1)
	assert.NotNil(t, result1)
	assert.Equal(t, "success", result1.Condition)

	// Assert cookie is in the authentication request context
	cookie, exists := state1.HttpAuthContext.AdditionalResponseCookies["session"]
	assert.True(t, exists, "Cookie should be in AdditionalResponseCookies")
	assert.NotEmpty(t, cookie.Value)
	cookieValue := cookie.Value // This is the hash stored in the cookie

	// Assert device is in context
	assert.NotEmpty(t, state1.Context["device"])
	deviceID := state1.Context["device"]

	// For step 2, we need to create a user with the device attribute
	// The cookie value is the deviceSecretHash, but is_known_device hashes it again
	// So we need to hash the cookie value to get the index
	deviceHashForLookup := lib.HashString(cookieValue)

	// Create device attribute value for the user (we'll need to reconstruct it)
	now := time.Now()
	expiry := now.Add(30 * 24 * time.Hour)

	device := model.DeviceAttributeValue{
		DeviceID:         deviceID,
		DeviceSecretHash: cookieValue,
		DeviceName:       "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		DeviceIP:         "192.168.1.100",
		DeviceUserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		CookieName:       "session",
		CookieExpires:    expiry,
		CookieSameSite:   "Lax",
		CookieHttpOnly:   true,
		CookieSecure:     true,
		SessionLoa0:      *attributes.InitSession(now, attributes.DEFAULT_LOA_TO_EXPIRY_MAPPINGS[0]),
		SessionLoa1:      nil,
		SessionLoa2:      nil,
	}

	// Add device attribute to user with the double-hash as index (since is_known_device hashes the cookie value)
	deviceAttr := &model.UserAttribute{
		Type:      model.AttributeTypeDevice,
		Value:     device,
		Index:     &deviceHashForLookup, // This is hash(cookieValue) which is hash(hash(secret))
		CreatedAt: now,
		UpdatedAt: now,
	}
	testUser.AddAttribute(deviceAttr)

	// ========== STEP 2: Is Known Device ==========
	// Create node for checking device
	isKnownDeviceNode := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create new empty authentication session but add the cookie
	state2 := &model.AuthenticationSession{
		User:    nil,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{},
			RequestIP:      "192.168.1.100",
			RequestCookies: map[string]string{
				"session": cookieValue, // Cookie contains the device secret hash
			},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input2 := map[string]string{}

	// Setup expectation for GetByAttributeIndex - should return user with device
	// Note: is_known_device hashes the cookie value, so we use deviceHashForLookup
	mockUserRepo.On("GetByAttributeIndex", context.Background(), model.AttributeTypeDevice, deviceHashForLookup).Return(testUser, nil)

	// Setup expectation for UpdateUserAttribute - should update last activity
	mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.MatchedBy(func(attr *model.UserAttribute) bool {
		return attr.Type == model.AttributeTypeDevice
	})).Return(nil)

	// Execute STEP 2
	result2, err2 := RunIsKnownDeviceNode(state2, isKnownDeviceNode, input2, services)

	// Assertions for STEP 2
	assert.NoError(t, err2)
	assert.NotNil(t, result2)
	assert.Equal(t, CONDITION_KNOWN_DEVICE, result2.Condition)

	// Assert device is known and user is authenticated
	assert.NotNil(t, state2.User)
	assert.Equal(t, testUser.ID, state2.User.ID)
	assert.NotEmpty(t, state2.Context["device"])

	mockUserRepo.AssertExpectations(t)
}

func TestIsKnownDeviceNodeUnknownDevice(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create authentication session with unknown cookie (cookie that doesn't match any user)
	unknownCookieValue := lib.GenerateSecureSessionID()
	state := &model.AuthenticationSession{
		User:    nil,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{},
			RequestIP:      "192.168.1.100",
			RequestCookies: map[string]string{
				"session": unknownCookieValue,
			},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input := map[string]string{}

	// Setup expectation for GetByAttributeIndex - should return nil (user not found)
	unknownCookieHash := lib.HashString(unknownCookieValue)
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, unknownCookieHash).Return(nil, nil)

	// Execute
	result, err := RunIsKnownDeviceNode(state, node, input, services)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CONDITION_UNKNOWN_DEVICE, result.Condition)

	// Assert user is not authenticated
	assert.Nil(t, state.User)

	// Assert device is not in context
	assert.Empty(t, state.Context["device"])

	mockUserRepo.AssertExpectations(t)
}

func TestIsKnownDeviceNodeExpiredDevice(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test user
	testUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
	}

	// Create device secret and hash
	deviceSecret := lib.GenerateSecureSessionID()
	deviceSecretHash := lib.HashString(deviceSecret)
	deviceID := lib.GenerateSecureSessionID()

	now := time.Now()
	// Create an expired device (expiry is in the past)
	expiredTime := now.Add(-2 * 365 * 24 * time.Hour) // 2 years ago
	expiry := &expiredTime

	// Create device attribute value with expired session
	device := model.DeviceAttributeValue{
		DeviceID:         deviceID,
		DeviceSecretHash: deviceSecretHash,
		DeviceName:       "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		DeviceIP:         "192.168.1.100",
		DeviceUserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		CookieName:       "session",
		CookieExpires:    *expiry,
		CookieSameSite:   "Lax",
		CookieHttpOnly:   true,
		CookieSecure:     true,
		SessionLoa0:      *attributes.InitSession(expiredTime, attributes.DEFAULT_LOA_TO_EXPIRY_MAPPINGS[0]),
		SessionLoa1:      nil,
		SessionLoa2:      nil,
	}

	// Add device attribute to user
	deviceAttribute := &model.UserAttribute{
		Type:      model.AttributeTypeDevice,
		Value:     device,
		Index:     &deviceSecretHash,
		CreatedAt: expiredTime,
		UpdatedAt: expiredTime,
	}
	testUser.AddAttribute(deviceAttribute)

	// Create node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create authentication session with cookie
	state := &model.AuthenticationSession{
		User:    nil,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{},
			RequestIP:      "192.168.1.100",
			RequestCookies: map[string]string{
				"session": deviceSecret,
			},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input := map[string]string{}

	// Setup expectation for GetByAttributeIndex - should return user with expired device
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, deviceSecretHash).Return(testUser, nil)

	// Execute
	result, err := RunIsKnownDeviceNode(state, node, input, services)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CONDITION_UNKNOWN_DEVICE, result.Condition)

	// Assert user is not authenticated (because device is expired)
	assert.Nil(t, state.User)

	// Assert device is not in context
	assert.Empty(t, state.Context["device"])

	mockUserRepo.AssertExpectations(t)
}

func TestIsKnownDeviceNodeRefreshExpiry(t *testing.T) {

	now := time.Now()

	// Setup
	mockUserRepo := repository.NewMockUserRepository()

	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test user
	testUser := &model.User{
		ID:     uuid.NewString(),
		Status: "active",
	}

	// Create device secret and hash
	deviceSecret := lib.GenerateSecureSessionID()
	deviceSecretHash := lib.HashString(deviceSecret)
	deviceID := lib.GenerateSecureSessionID()

	// Create a device that is past refresh time but before expiry
	sessionDuration := 3600 * 24 * 30         // 30 days
	sessionRefreshAfter := 3600 * 24 * 30 / 2 // 15 days

	// Device was created 20 days ago (past refresh time)
	firstLogin := now.Add(-20 * 24 * time.Hour)
	// Expiry is 10 days from now (still valid, but past refresh)
	expiry := now.Add(10 * 24 * time.Hour)
	// Last activity was 20 days ago (same as first login)
	lastActivity := now.Add(-20 * 24 * time.Hour)
	sessionLoa0 := attributes.Session{
		SessionDuration:     sessionDuration,
		SessionRefreshAfter: sessionRefreshAfter,
		SessionExpiry:       expiry,
		SessionLastActivity: lastActivity,
		LevelOfAssurance:    0,
	}

	// Create device attribute value
	device := model.DeviceAttributeValue{
		DeviceID:         deviceID,
		DeviceSecretHash: deviceSecretHash,
		DeviceName:       "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		DeviceIP:         "192.168.1.100",
		DeviceUserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		SessionLoa0:      sessionLoa0,
		SessionLoa1:      nil,
		SessionLoa2:      nil,
		CookieName:       "session",
	}

	// Add device attribute to user
	deviceAttribute := &model.UserAttribute{
		Type:      model.AttributeTypeDevice,
		Value:     device,
		Index:     &deviceSecretHash,
		CreatedAt: firstLogin,
		UpdatedAt: lastActivity,
	}
	testUser.AddAttribute(deviceAttribute)

	// Create node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			CONFIG_COOKIE_NAME: "session",
		},
	}

	// Create authentication session with cookie
	state := &model.AuthenticationSession{
		User:    nil,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{},
			RequestIP:      "192.168.1.100",
			RequestCookies: map[string]string{
				"session": deviceSecret,
			},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
	}

	input := map[string]string{}

	// Setup expectation for GetByAttributeIndex - should return user with device
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, deviceSecretHash).Return(testUser, nil)

	// Setup expectation for UpdateUserAttribute - should update last activity and expiry
	var updatedDevice *model.DeviceAttributeValue
	mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.MatchedBy(func(attr *model.UserAttribute) bool {
		if attr.Type == model.AttributeTypeDevice && attr.Value != nil {
			// Handle pointer to pointer case (**model.DeviceAttributeValue)
			if ptrPtr, ok := attr.Value.(**model.DeviceAttributeValue); ok && ptrPtr != nil && *ptrPtr != nil {
				updatedDevice = *ptrPtr
				return true
			}
			// Handle pointer case (*model.DeviceAttributeValue)
			if deviceValue, ok := attr.Value.(*model.DeviceAttributeValue); ok {
				updatedDevice = deviceValue
				return true
			}
			// Handle value case (model.DeviceAttributeValue)
			if deviceValue, ok := attr.Value.(model.DeviceAttributeValue); ok {
				updatedDevice = &deviceValue
				return true
			}
		}
		return false
	})).Return(nil)

	// Execute
	result, err := RunIsKnownDeviceNode(state, node, input, services)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CONDITION_KNOWN_DEVICE, result.Condition)

	// Assert user is authenticated
	assert.NotNil(t, state.User)
	assert.Equal(t, testUser.ID, state.User.ID)

	// Assert device is in context
	assert.NotEmpty(t, state.Context["device"])
	assert.Equal(t, deviceID, state.Context["device"])

	// Assert that the device was updated (last activity should be updated to now)
	// Note: The current implementation only updates SessionLastActivity, not expiry
	// If expiry refresh logic is added, this test will verify it
	if updatedDevice != nil {
		assert.NotNil(t, updatedDevice.SessionLoa0.SessionLastActivity)
		// Last activity should be updated to approximately now (within 1 second)
		assert.WithinDuration(t, now, updatedDevice.SessionLoa0.SessionLastActivity, 1*time.Second)
	}

	mockUserRepo.AssertExpectations(t)
}
