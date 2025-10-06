package node_telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEmailSender is a mock implementation of the email sender
type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendEmail(subject, body, recipientEmail, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail string) error {
	args := m.Called(subject, body, recipientEmail, smtpServer, smtpPort, smtpUsername, smtpPassword, smtpSenderEmail)
	return args.Error(0)
}

// generateValidTestAuthResult creates a valid test auth result with proper hash for the given token
func generateValidTestAuthResult(botToken string, authDate int64) string {
	// Create test data
	authData := map[string]interface{}{
		"id":         int64(6745731120),
		"first_name": "Luca",
		"username":   "WhoIsLuca",
		"photo_url":  "https://t.me/i/userpic/ABC",
		"auth_date":  authDate,
	}

	// Create data map for hash calculation (same as in verifyTelegramHashWithTime)
	dataMap := map[string]string{
		"id":         strconv.FormatInt(authData["id"].(int64), 10),
		"first_name": authData["first_name"].(string),
		"username":   authData["username"].(string),
		"photo_url":  authData["photo_url"].(string),
		"auth_date":  strconv.FormatInt(authData["auth_date"].(int64), 10),
	}

	// Create data check array (key=value pairs)
	var dataCheckArr []string
	for key, value := range dataMap {
		dataCheckArr = append(dataCheckArr, key+"="+value)
	}

	// Sort the array
	sort.Strings(dataCheckArr)

	// Join with newlines
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Create secret key (SHA256 hash of bot token)
	secretKey := sha256.Sum256([]byte(botToken))

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	calculatedHash := fmt.Sprintf("%x", h.Sum(nil))

	// Add the hash to the auth data
	authData["hash"] = calculatedHash

	// Convert to JSON and base64 encode
	jsonData, _ := json.Marshal(authData)
	return base64.StdEncoding.EncodeToString(jsonData)
}

// getTestTelegramCredentials returns test tokens and auth results for testing
func getTestTelegramCredentials() (string, string) {
	// Test bot token (not a real token)
	testToken := "1234567890:TestBotTokenForTestingPurposesOnly"

	// Generate a valid auth result with proper hash for the test token
	// Use a recent auth date to avoid outdated issues
	authDate := time.Now().Unix() - 60 // 1 minute ago
	testAuthResult := generateValidTestAuthResult(testToken, authDate)

	return testToken, testAuthResult
}

// getOutdatedTestTelegramCredentials returns test tokens and auth results that are outdated for testing
func getOutdatedTestTelegramCredentials() (string, string) {
	// Test bot token (not a real token)
	testToken := "1234567890:TestBotTokenForTestingPurposesOnly"

	// Generate a valid auth result with proper hash for the test token
	// Use an old auth date to test outdated functionality
	authDate := time.Now().Unix() - 1000 // 1000 seconds ago (more than 15 minutes)
	testAuthResult := generateValidTestAuthResult(testToken, authDate)

	return testToken, testAuthResult
}

// runTelegramLoginNodeWithTime is a test version that accepts a custom time
func runTelegramLoginNodeWithTime(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories, currentTime time.Time) (*model.NodeResult, error) {
	botToken := node.CustomConfig["telegram_bottoken"]

	// Get the bot id from the bot token
	botId := strings.Split(botToken, ":")[0]
	if botId == "" {
		return model.NewNodeResultWithError(fmt.Errorf("botToken is not valid"))
	}

	if input["tgAuthResult"] != "" {
		// Parse the result with custom time
		authResult, err := parseTelegramAuthResultWithTime(input["tgAuthResult"], botToken, currentTime)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}

		telegramUserID := strconv.FormatInt(authResult.ID, 10)

		// Create Telegram attribute value
		telegramAttributeValue := model.TelegramAttributeValue{
			TelegramUserID:    telegramUserID,
			TelegramUsername:  authResult.Username,
			TelegramFirstName: authResult.FirstName,
			TelegramPhotoURL:  authResult.PhotoURL,
			TelegramAuthDate:  authResult.AuthDate,
		}

		// Store the telegram attribute in the context
		telegramAttributeJSON, _ := json.Marshal(telegramAttributeValue)
		state.Context["telegram"] = string(telegramAttributeJSON)

		// Check if the user exists using the new attribute system
		dbUser, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeTelegram, telegramUserID)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}

		// if the user exists we put it to the context and return
		if dbUser != nil {
			state.User = dbUser
			return model.NewNodeResultWithCondition("existing-user")
		}

		// If the create user option is enabled we create a new user if it doesn't exist
		if node.CustomConfig["telegram_create_user"] == "true" {
			user := &model.User{
				ID:     uuid.NewString(),
				Status: "active",
			}

			// Add the telegram attribute to the user
			user.AddAttribute(&model.UserAttribute{
				Index: lib.StringPtr(telegramUserID),
				Type:  model.AttributeTypeTelegram,
				Value: telegramAttributeValue,
			})

			// Create the user
			err := services.UserRepo.Create(context.Background(), user)
			if err != nil {
				return model.NewNodeResultWithError(err)
			}

			state.User = user
			return model.NewNodeResultWithCondition("new-user")
		}

		// If we should not create a user just set the telegram fields to the context
		state.User = &model.User{
			Status: "active",
		}

		// Add the telegram attribute to the user
		state.User.AddAttribute(&model.UserAttribute{
			Index: lib.StringPtr(telegramUserID),
			Type:  model.AttributeTypeTelegram,
			Value: telegramAttributeValue,
		})

		// Return new-user condition even when not creating the user
		return model.NewNodeResultWithCondition("new-user")
	}

	// Get the last history entry
	lastHistory := ""
	if len(state.History) > 0 {
		lastHistory = state.History[len(state.History)-1]
	}
	if strings.HasPrefix(lastHistory, "telegramLogin:prompted") {
		return model.NewNodeResultWithPrompts(map[string]string{
			"tgAuthResult": "base64",
		})
	}

	// Otherwise we redirect to Telegram for the login flow
	origin := getOrigin(state.LoginUriBase)

	// Assemble the url with query params using net/url
	baseUrl := "https://oauth.telegram.org/auth"
	params := url.Values{}
	params.Set("bot_id", botId)
	params.Set("origin", origin)
	params.Set("return_to", state.LoginUriBase+"?callback=telegram")

	if node.CustomConfig["telegram_request_write_access"] == "true" {
		params.Set("request_access", "write")
	}

	telegramUrl := baseUrl + "?" + params.Encode()

	return model.NewNodeResultWithPrompts(map[string]string{
		"__redirect":   telegramUrl,
		"tgAuthResult": "base64",
	})
}

func TestParseTelegramAuthResult(t *testing.T) {
	tgToken, tgAuthResult := getTestTelegramCredentials()

	// Use a fixed time for testing to ensure consistent results
	testTime := time.Unix(1753279000, 0) // Slightly after the auth_date in the test data

	authResult, err := parseTelegramAuthResultWithTime(tgAuthResult, tgToken, testTime)
	if err != nil {
		t.Fatalf("Failed to parse telegram auth result: %v", err)
	}

	fmt.Printf("Test 1 - Auth Result: %+v\n", authResult)
}

func TestParseTelegramAuthResultWithDifferentToken(t *testing.T) {
	tgToken := "1234567890:DifferentBotTokenForTestingPurposes"
	// This test data has a hash that doesn't match the token, which should fail verification
	tgAuthResult := "eyJpZCI6MTIzNDU2Nzg5MCwiZmlyc3RfbmFtZSI6IkpvaG4iLCJ1c2VybmFtZSI6ImpvaG5kb2UiLCJwaG90b191cmwiOiJodHRwczpcL1wvdC5tZVwvaVwvdXNlcnBpY1wvMTIzXC9waG90by5qcGciLCJhdXRoX2RhdGUiOjE3NTMyNzg5ODcsImhhc2giOiJhYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5ejEyMzQ1Njc4OTAiLCJfX2V4dHJhX2ZpZWxkIjoiZXh0cmFfdmFsdWUifQ=="

	// Use a fixed time for testing to ensure consistent results
	testTime := time.Unix(1753279000, 0) // Slightly after the auth_date in the test data

	_, err := parseTelegramAuthResultWithTime(tgAuthResult, tgToken, testTime)
	if err == nil {
		t.Fatalf("Expected error for invalid hash, but got none")
	}

	if err.Error() != "hash verification failed: hash verification failed - data is not from Telegram" {
		t.Fatalf("Expected 'data is not from Telegram' error, but got: %v", err)
	}

	fmt.Println("Test 2 - Correctly detected invalid hash for different bot token")
}

func TestParseTelegramAuthResultInvalidHash(t *testing.T) {
	tgToken, _ := getTestTelegramCredentials()
	// Modified the original test data with an invalid hash
	tgAuthResult := "eyJpZCI6Njc0NTczMTEyMCwiZmlyc3RfbmFtZSI6Ikx1Y2EiLCJ1c2VybmFtZSI6Ik1pZ2h0VGhpc0JlTHVjYSIsInBob3RvX3VybCI6Imh0dHBzOlwvXC90Lm1lXC9pXC91c2VycGljXC8zMjBcL3AxU1NkYkRvUjFvTnN6MEdHLUEwUlFHSkFhbmZaWkp5UEpVNFN1VEo5V01aQ2dYY0laLUtsYmZJc3pMQ2tfZlkuanBnIiwiYXV0aF9kYXRlIjoxNzUzMjc4OTg2LCJoYXNoIjoiaW52YWxpZGhhc2h0ZXN0In0="

	// Use a fixed time for testing to ensure consistent results
	testTime := time.Unix(1753279000, 0) // Slightly after the auth_date in the test data

	_, err := parseTelegramAuthResultWithTime(tgAuthResult, tgToken, testTime)
	if err == nil {
		t.Fatalf("Expected error for invalid hash, but got none")
	}

	if err.Error() != "hash verification failed: hash verification failed - data is not from Telegram" {
		t.Fatalf("Expected 'data is not from Telegram' error, but got: %v", err)
	}

	fmt.Println("Test 4 - Correctly detected invalid hash")
}

func TestParseTelegramAuthResultOutdated(t *testing.T) {
	tgToken, tgAuthResult := getOutdatedTestTelegramCredentials()

	// Use current time to test outdated check
	testTime := time.Now()

	_, err := parseTelegramAuthResultWithTime(tgAuthResult, tgToken, testTime)
	if err == nil {
		t.Fatalf("Expected error for outdated auth data, but got none")
	}

	if err.Error() != "hash verification failed: data is outdated" {
		t.Fatalf("Expected 'data is outdated' error, but got: %v", err)
	}

	fmt.Println("Test 3 - Correctly detected outdated auth data")
}

func TestRunTelegramLoginNodeExistingUser(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	mockEmailSender := &MockEmailSender{}

	services := &model.Repositories{
		UserRepo:    mockUserRepo,
		EmailSender: mockEmailSender,
	}

	// Create test data
	tgToken, tgAuthResult := getTestTelegramCredentials()

	// Create existing user with Telegram attribute
	existingUser := &model.User{
		ID:     "existing-user-id",
		Status: "active",
	}

	// Add Telegram attribute
	telegramAttributeValue := model.TelegramAttributeValue{
		TelegramUserID:    "6745731120",
		TelegramUsername:  "WhoIsLuca",
		TelegramFirstName: "Luca",
		TelegramPhotoURL:  "https://t.me/i/userpic/ABC",
		TelegramAuthDate:  time.Now().Unix(),
	}

	existingUser.AddAttribute(&model.UserAttribute{
		Index: lib.StringPtr("6745731120"),
		Type:  model.AttributeTypeTelegram,
		Value: telegramAttributeValue,
	})

	// Setup expectations
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeTelegram, "6745731120").Return(existingUser, nil)

	// Create node and state
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"telegram_bottoken": tgToken,
		},
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{},
	}

	input := map[string]string{
		"tgAuthResult": tgAuthResult,
	}

	// Use a fixed time for testing
	testTime := time.Now()

	// Execute
	result, err := runTelegramLoginNodeWithTime(state, node, input, services, testTime)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "existing-user", result.Condition)
	assert.Equal(t, existingUser, state.User)

	mockUserRepo.AssertExpectations(t)
}

func TestRunTelegramLoginNodeNewUser(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	mockEmailSender := &MockEmailSender{}

	services := &model.Repositories{
		UserRepo:    mockUserRepo,
		EmailSender: mockEmailSender,
	}

	// Create test data
	tgToken, tgAuthResult := getTestTelegramCredentials()

	// Setup expectations - user doesn't exist initially
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeTelegram, "6745731120").Return(nil, nil)

	// Expect user creation
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *model.User) bool {
		// Check that the user has the correct Telegram attribute
		telegramAttr, _, err := model.GetAttribute[model.TelegramAttributeValue](user, model.AttributeTypeTelegram)
		return err == nil && telegramAttr != nil &&
			telegramAttr.TelegramUsername == "WhoIsLuca" &&
			telegramAttr.TelegramPhotoURL == "https://t.me/i/userpic/ABC" &&
			telegramAttr.TelegramUserID == "6745731120"
	})).Return(nil)

	// Create node and state
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"telegram_bottoken":    tgToken,
			"telegram_create_user": "true",
		},
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{},
	}

	input := map[string]string{
		"tgAuthResult": tgAuthResult,
	}

	// Use a fixed time for testing
	testTime := time.Now()

	// Execute
	result, err := runTelegramLoginNodeWithTime(state, node, input, services, testTime)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-user", result.Condition)
	assert.NotNil(t, state.User)

	// Verify that the Telegram attribute was created correctly
	telegramAttr, _, err := model.GetAttribute[model.TelegramAttributeValue](state.User, model.AttributeTypeTelegram)
	assert.NoError(t, err)
	assert.NotNil(t, telegramAttr)
	assert.Equal(t, "WhoIsLuca", telegramAttr.TelegramUsername)
	assert.Equal(t, "https://t.me/i/userpic/ABC", telegramAttr.TelegramPhotoURL)
	assert.Equal(t, "6745731120", telegramAttr.TelegramUserID)

	mockUserRepo.AssertExpectations(t)
}

func TestRunTelegramLoginNodeNoCreateUser(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	mockEmailSender := &MockEmailSender{}

	services := &model.Repositories{
		UserRepo:    mockUserRepo,
		EmailSender: mockEmailSender,
	}

	// Create test data
	tgToken, tgAuthResult := getTestTelegramCredentials()

	// Setup expectations - user doesn't exist initially
	mockUserRepo.On("GetByAttributeIndex", mock.Anything, model.AttributeTypeTelegram, "6745731120").Return(nil, nil)

	// Create node and state
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"telegram_bottoken":    tgToken,
			"telegram_create_user": "false", // Don't create user
		},
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{},
	}

	input := map[string]string{
		"tgAuthResult": tgAuthResult,
	}

	// Use a fixed time for testing
	testTime := time.Now()

	// Execute
	result, err := runTelegramLoginNodeWithTime(state, node, input, services, testTime)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-user", result.Condition) // Should still return new-user even without creating
	assert.NotNil(t, state.User)

	// Verify that the Telegram attribute was created correctly
	telegramAttr, _, err := model.GetAttribute[model.TelegramAttributeValue](state.User, model.AttributeTypeTelegram)
	assert.NoError(t, err)
	assert.NotNil(t, telegramAttr)
	assert.Equal(t, "WhoIsLuca", telegramAttr.TelegramUsername)
	assert.Equal(t, "https://t.me/i/userpic/ABC", telegramAttr.TelegramPhotoURL)
	assert.Equal(t, "6745731120", telegramAttr.TelegramUserID)

	mockUserRepo.AssertExpectations(t)
}

func TestRunTelegramLoginNodeRedirect(t *testing.T) {
	// Setup
	mockUserRepo := repository.NewMockUserRepository()
	mockEmailSender := &MockEmailSender{}

	services := &model.Repositories{
		UserRepo:    mockUserRepo,
		EmailSender: mockEmailSender,
	}

	// Create node and state
	node := &model.GraphNode{
		CustomConfig: map[string]string{
			"telegram_bottoken": "1234567890:TestBotTokenForTestingPurposesOnly",
		},
	}

	state := &model.AuthenticationSession{
		Context: map[string]string{},
	}

	input := map[string]string{
		"tgAuthResult": "", // No auth result, should redirect
	}

	// Execute
	result, err := RunTelegramLoginNode(state, node, input, services)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Prompts["__redirect"])
	assert.Contains(t, result.Prompts["__redirect"], "https://oauth.telegram.org/auth")
	assert.Contains(t, result.Prompts["__redirect"], "bot_id=1234567890")
}
