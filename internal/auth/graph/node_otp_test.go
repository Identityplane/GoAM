package graph

import (
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/gianlucafrei/GoAM/internal/auth/repository"
	"github.com/gianlucafrei/GoAM/internal/model"

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
	// Create test user
	testUser := &model.User{
		ID:                     uuid.NewString(),
		Email:                  "test@example.com",
		FailedLoginAttemptsMFA: 0,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Create mock repository
	mockUserRepo := new(MockUserRepository)
	services := &repository.Repositories{
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
	}

	// Test 1: Initial state - should return prompt
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
	result, err := RunEmailOTPNode(session, node, map[string]string{}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])
	assert.NotEmpty(t, session.Context["email_otp"])

	// Store the generated OTP for later use
	generatedOTP := session.Context["email_otp"]

	// Test 2: Wrong OTP - should return prompt and increase counter
	mockUserRepo.On("Update", mock.Anything, testUser).Return(nil)
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": "000000"}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])
	assert.Equal(t, 1, testUser.FailedLoginAttemptsMFA)

	// Test 3: Correct OTP - should return success
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": generatedOTP}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Condition)
	assert.Equal(t, 0, testUser.FailedLoginAttemptsMFA) // counter should be reset to 0

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
}
