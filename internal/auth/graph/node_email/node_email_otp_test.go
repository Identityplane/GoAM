package node_email

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateOTP(t *testing.T) {
	// Run multiple tests to ensure consistency
	for i := 0; i < 100; i++ {
		otp := generateOTP()

		// Test 1: Check if OTP is exactly 6 digits
		if len(otp) != 6 {
			t.Errorf("OTP length is not 6 digits, got length %d: %s", len(otp), otp)
		}

		// Test 2: Check if OTP contains only digits
		matched, err := regexp.MatchString(`^\d{6}$`, otp)
		if err != nil {
			t.Fatalf("Error in regex matching: %v", err)
		}
		if !matched {
			t.Errorf("OTP contains non-digit characters: %s", otp)
		}

		// Test 3: Check if OTP is within valid range (000000-999999)
		num, err := strconv.Atoi(otp)
		if err != nil {
			t.Errorf("Failed to convert OTP to number: %v", err)
		}
		if num < 0 || num > 999999 {
			t.Errorf("OTP is outside valid range (0-999999): %d", num)
		}
	}
}

func TestRunEmailOTPNode(t *testing.T) {
	// Create test user with email attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add email attribute with OTP fields
	emailAttrValue := model.EmailAttributeValue{
		Email:             "test@example.com",
		Verified:          false,
		VerifiedAt:        nil,
		OtpFailedAttempts: 0,
		OtpLocked:         false,
	}
	testUser.AddAttribute(&model.UserAttribute{
		ID:    uuid.NewString(),
		Type:  model.AttributeTypeEmail,
		Index: "test@example.com",
		Value: emailAttrValue,
	})

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"mfa_max_attempts": "3",
		},
	}

	// Create test session
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"email": "test@example.com",
		},
		User: testUser,
	}

	// Test 1: Initial state - should return prompt
	result, err := RunEmailOTPNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])
	assert.NotEmpty(t, session.Context["email_otp"])

	// Store the generated OTP for later use
	generatedOTP := session.Context["email_otp"]

	// Test 2: Wrong OTP - should return prompt and increase counter
	mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.AnythingOfType("*model.UserAttribute")).Return(nil)
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": "000000"}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])

	// Verify failed attempts increased
	emailAttr, _, err := model.GetAttribute[model.EmailAttributeValue](testUser, model.AttributeTypeEmail)
	assert.NoError(t, err)
	assert.NotNil(t, emailAttr)
	assert.Equal(t, 1, emailAttr.OtpFailedAttempts)

	// Test 3: Correct OTP - should return success
	mockUserRepo.On("UpdateUserAttribute", mock.Anything, mock.AnythingOfType("*model.UserAttribute")).Return(nil)
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": generatedOTP}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)

	// Verify failed attempts were reset and email was verified
	emailAttr, _, err = model.GetAttribute[model.EmailAttributeValue](testUser, model.AttributeTypeEmail)
	assert.NoError(t, err)
	assert.NotNil(t, emailAttr)
	assert.Equal(t, 0, emailAttr.OtpFailedAttempts) // counter should be reset to 0
	assert.False(t, emailAttr.OtpLocked)            // should be unlocked
	assert.True(t, emailAttr.Verified)              // should be verified
	assert.NotNil(t, emailAttr.VerifiedAt)          // verification timestamp should be set

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}

func TestRunEmailOTPNode_NoUser(t *testing.T) {
	// Test OTP functionality without a user (for onboarding scenarios)

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"mfa_max_attempts": "3",
		},
	}

	// Create test session without user
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"email": "newuser@example.com",
		},
		User: nil, // No user for onboarding
	}

	// Mock the GetByAttributeIndex call that TryLoadUserFromContext will make
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeEmail, "newuser@example.com").Return(nil, nil)

	// Test 1: Initial state - should return prompt
	result, err := RunEmailOTPNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])
	assert.NotEmpty(t, session.Context["email_otp"])

	// Store the generated OTP for later use
	generatedOTP := session.Context["email_otp"]

	// Test 2: Wrong OTP - should return prompt (no counter increase since no user)
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": "000000"}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])

	// Test 3: Correct OTP - should return success
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": generatedOTP}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}

func TestRunEmailOTPNode_AccountLocked(t *testing.T) {
	// Test OTP functionality with a locked account

	// Create test user with locked email attribute
	testUser := &model.User{
		ID:        uuid.NewString(),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add email attribute that is locked
	emailAttrValue := model.EmailAttributeValue{
		Email:             "locked@example.com",
		Verified:          false,
		VerifiedAt:        nil,
		OtpFailedAttempts: 3,    // At max attempts
		OtpLocked:         true, // Locked
	}
	testUser.AddAttribute(&model.UserAttribute{
		ID:    uuid.NewString(),
		Type:  model.AttributeTypeEmail,
		Index: "locked@example.com",
		Value: emailAttrValue,
	})

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	services := &model.Repositories{
		UserRepo: mockUserRepo,
	}

	// Create test node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"mfa_max_attempts": "3",
		},
	}

	// Create test session
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"email": "locked@example.com",
		},
		User: testUser,
	}

	// Test: Should fail silently when account is locked (no OTP sent)
	result, err := RunEmailOTPNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])

	// OTP should still be generated but not sent via email
	assert.NotEmpty(t, session.Context["email_otp"])

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}
