package node_email

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/pkg/model"
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

func TestEmailOTP_NoUserInContext(t *testing.T) {
	// Create test user with email attribute

	// Create mock repository
	mockUserRepo := repository.NewMockUserRepository()
	mockEmailSender := repository.NewMockEmailSender()
	services := &model.Repositories{
		UserRepo:    mockUserRepo,
		EmailSender: mockEmailSender,
	}

	// Create test node
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			EMAIL_OTP_OPTION_INIT_USER: "true",
		},
	}

	// Create test session
	session := &model.AuthenticationSession{
		Context: map[string]string{
			"email": "test@example.com",
		},
	}

	// Mock SendEmail call for initial OTP generation
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeEmail, "test@example.com").Return(nil, nil)
	mockEmailSender.On("SendEmail", mock.AnythingOfType("*model.SendEmailParams")).Return(nil)

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
	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": "000000"}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts)
	assert.Equal(t, "number", result.Prompts["otp"])

	result, err = RunEmailOTPNode(session, node, map[string]string{"otp": generatedOTP}, services)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success-unkown-email", result.Condition)

	// Verify failed attempts were reset and email was verified
	user := session.User
	emailAttr, _, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
	assert.NoError(t, err)
	assert.NotNil(t, emailAttr)
	assert.Equal(t, 0, emailAttr.OtpFailedAttempts) // counter should be reset to 0
	assert.False(t, emailAttr.OtpLocked)            // should be unlocked
	assert.True(t, emailAttr.Verified)              // should be verified
	assert.NotNil(t, emailAttr.VerifiedAt)          // verification timestamp should be set

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}
