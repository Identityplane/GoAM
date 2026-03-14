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
		CustomConfig: map[string]string{},
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

	// Get device ID from context
	deviceID := state.Context["device"]
	assert.NotEmpty(t, deviceID)
	cookieName := "device_" + deviceID[:8]

	// Assert cookie is in the authentication request context
	cookie, exists := state.HttpAuthContext.AdditionalResponseCookies[cookieName]
	assert.True(t, exists, "Cookie should be in AdditionalResponseCookies")
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, cookieName, cookie.Name)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)

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
	cookieName := "device_" + deviceID[:8]

	now := time.Now()
	expiry := now.Add(30 * 24 * time.Hour) // 30 days

	// Create device attribute value
	device := model.DeviceAttributeValue{
		DeviceID:         deviceID,
		DeviceSecretHash: deviceSecretHash,
		DeviceName:       "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		DeviceIP:         "192.168.1.100",
		DeviceUserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		CookieName:       cookieName,
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
		CustomConfig: map[string]string{},
	}

	// Create empty authentication session with cookie
	state := &model.AuthenticationSession{
		User:    nil,
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestHeaders: map[string]string{},
			RequestIP:      "192.168.1.100",
			RequestCookies: map[string]string{
				cookieName: deviceSecret, // Cookie contains the device secret, not the hash
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
	addDeviceNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

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

	// Setup expectation for CreateUserAttribute
	mockUserRepo.On("CreateUserAttribute", mock.Anything, mock.MatchedBy(func(attr *model.UserAttribute) bool {
		return attr.Type == model.AttributeTypeDevice
	})).Return(nil)

	// Execute STEP 1
	RunAddKnownDeviceNode(state1, addDeviceNode, nil, services)

	deviceID := state1.Context["device"]
	cookieName := "device_" + deviceID[:8]
	cookie := state1.HttpAuthContext.AdditionalResponseCookies[cookieName]
	deviceSecret := cookie.Value
	deviceSecretHash := lib.HashString(deviceSecret)

	// For step 2, we need the double hash for lookup
	deviceHashForLookup := lib.HashString(deviceSecretHash)

	// Create device attribute value
	now := time.Now()
	device := model.DeviceAttributeValue{
		DeviceID:         deviceID,
		DeviceSecretHash: deviceSecretHash,
		DeviceName:       "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		DeviceIP:         "192.168.1.100",
		DeviceUserAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
		CookieName:       cookieName,
		CookieExpires:    now.Add(30 * 24 * time.Hour),
		SessionLoa0:      *attributes.InitSession(now, attributes.DEFAULT_LOA_TO_EXPIRY_MAPPINGS[0]),
	}

	deviceAttr := &model.UserAttribute{
		Type:      model.AttributeTypeDevice,
		Value:     device,
		Index:     &deviceHashForLookup,
		CreatedAt: now,
		UpdatedAt: now,
	}
	testUser.AddAttribute(deviceAttr)

	// ========== STEP 2: Is Known Device ==========
	isKnownDeviceNode := &model.GraphNode{
		CustomConfig: map[string]string{},
	}

	state2 := &model.AuthenticationSession{
		HttpAuthContext: &model.HttpAuthContext{
			RequestIP: "192.168.1.100",
			RequestCookies: map[string]string{
				cookieName: deviceSecretHash,
			},
			AdditionalResponseCookies: make(map[string]http.Cookie),
		},
		Context: map[string]string{},
	}

	mockUserRepo.On("GetByAttributeIndex", context.Background(), model.AttributeTypeDevice, deviceHashForLookup).Return(testUser, nil)
	mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.Anything).Return(nil)

	// Execute STEP 2
	result2, err2 := RunIsKnownDeviceNode(state2, isKnownDeviceNode, nil, services)

	assert.NoError(t, err2)
	assert.Equal(t, CONDITION_KNOWN_DEVICE, result2.Condition)
	assert.Equal(t, testUser.ID, state2.User.ID)
}

func TestIsKnownDeviceNodeMultipleDevicesSameUser(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{UserRepo: mockUserRepo}

	secret1 := "secret1"
	secret2 := "secret2"
	hash1 := lib.HashString(secret1)
	hash2 := lib.HashString(secret2)

	now := time.Now()
	testUser1 := &model.User{ID: "user1", Status: "active"}
	
	// Device 1: LOA 0
	testUser1.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeDevice,
		Value: model.DeviceAttributeValue{
            DeviceID: "device1", 
            SessionLoa0: attributes.Session{LevelOfAssurance: 0, SessionExpiry: now.Add(time.Hour)},
        },
		Index: &hash1,
	})
	// Device 2: LOA 1
	testUser1.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeDevice,
		Value: model.DeviceAttributeValue{
            DeviceID: "device2", 
            SessionLoa0: attributes.Session{LevelOfAssurance: 0, SessionExpiry: now.Add(time.Hour)},
            SessionLoa1: &attributes.Session{LevelOfAssurance: 1, SessionExpiry: now.Add(time.Hour)},
        },
		Index: &hash2,
	})

	state := &model.AuthenticationSession{
		HttpAuthContext: &model.HttpAuthContext{
			RequestCookies: map[string]string{
				"device_12345678": secret1,
				"device_87654321": secret2,
			},
            AdditionalResponseCookies: make(map[string]http.Cookie),
		},
        Context: make(map[string]string),
	}

	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, hash1).Return(testUser1, nil)
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, hash2).Return(testUser1, nil)
    mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.Anything).Return(nil)

	// Execute
	result, err := RunIsKnownDeviceNode(state, &model.GraphNode{}, nil, services)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, CONDITION_KNOWN_DEVICE, result.Condition)
    assert.Equal(t, "1", state.Context["loa"]) // Should pick LOA 1
    assert.Equal(t, "device2", state.Context["device"]) // Should pick device2
}

func TestIsKnownDeviceNodeMultipleDevices(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{UserRepo: mockUserRepo}

	secret1 := "secret1"
	secret2 := "secret2"
	hash1 := lib.HashString(secret1)
	hash2 := lib.HashString(secret2)

	testUser1 := &model.User{ID: "user1", Status: "active"}
	testUser1.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeDevice,
		Value: model.DeviceAttributeValue{DeviceID: "device1"},
		Index: &hash1,
	})
	testUser2 := &model.User{ID: "user2", Status: "active"}
	testUser2.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeDevice,
		Value: model.DeviceAttributeValue{DeviceID: "device2"},
		Index: &hash2,
	})

	state := &model.AuthenticationSession{
		HttpAuthContext: &model.HttpAuthContext{
			RequestCookies: map[string]string{
				"device_12345678": secret1,
				"device_87654321": secret2,
			},
		},
	}

	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, hash1).Return(testUser1, nil)
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, hash2).Return(testUser2, nil)

	// Execute
	result, err := RunIsKnownDeviceNode(state, &model.GraphNode{}, nil, services)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, CONDITION_MULTIPLE_DEVICES, result.Condition)
}

func TestListAvailableUsersNode(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{UserRepo: mockUserRepo}

	id1 := "user1"
	secret1 := "secret1"
	hash1 := lib.HashString(secret1)
	deviceID1 := "deviceUUID1"
	cookieName := "device_" + deviceID1[:8]

	user1 := &model.User{ID: id1, Status: "active"}
	user1.AddAttribute(&model.UserAttribute{
		Type:  model.AttributeTypeDevice,
		Value: model.DeviceAttributeValue{DeviceID: deviceID1, DeviceSecretHash: hash1, CookieName: cookieName},
		Index: &hash1,
	})

	state := &model.AuthenticationSession{
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestCookies: map[string]string{
				cookieName: secret1,
			},
		},
	}

	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeDevice, hash1).Return(user1, nil)

	// Execute - Get list
	result, err := RunListAvailableUsersNode(state, nil, nil, services)
	assert.NoError(t, err)
	assert.Contains(t, result.Prompts["users"], id1)

	// Execute - Select user
	result, err = RunListAvailableUsersNode(state, nil, map[string]string{"userid": id1}, services)
	assert.NoError(t, err)
	assert.Equal(t, "success", result.Condition)
	assert.Equal(t, id1, state.User.ID)
	assert.Equal(t, deviceID1, state.Context["device"])
}

func TestListAvailableUsersNodeNone(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{UserRepo: mockUserRepo}

	state := &model.AuthenticationSession{
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestCookies: map[string]string{},
		},
	}

	// Execute
	result, err := RunListAvailableUsersNode(state, nil, nil, services)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "none", result.Condition)
}

func TestListAvailableUsersNodeAnotherAction(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{UserRepo: mockUserRepo}

	state := &model.AuthenticationSession{
		Context: map[string]string{},
		HttpAuthContext: &model.HttpAuthContext{
			RequestCookies: map[string]string{},
		},
	}

	// Execute
	result, err := RunListAvailableUsersNode(state, nil, map[string]string{"action": "another"}, services)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "none", result.Condition)
}
