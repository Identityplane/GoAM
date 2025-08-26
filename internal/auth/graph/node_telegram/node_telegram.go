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
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
)

const (
	// TelegramAuthOutdatedThreshold is the maximum age of Telegram auth data in seconds (15 minutes)
	TelegramAuthOutdatedThreshold = 15 * 60
	TelegramProviderString        = "telegram"
)

var TelegramLoginNode = &model.NodeDefinition{
	Name:                 "telegramLogin",
	PrettyName:           "Telegram Login",
	Description:          "Handles Telegram authentication flow that redirects to Telegram and then back to the app",
	Category:             "Social Login",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{""},
	PossiblePrompts:      map[string]string{"__redirect": "url", "tgAuthResult": "base64"},
	OutputContext:        []string{},
	PossibleResultStates: []string{"existing-user", "new-user", "failure"},
	CustomConfigOptions: map[string]string{
		"botToken":           "The bot token of the Telegram bot",
		"requestWriteAccess": "Whether to request write access to the user's Telegram account",
		"createUser":         "Whether to create a user if they don't exist",
	},
	Run: RunTelegramLoginNode,
}

func RunTelegramLoginNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	botToken := node.CustomConfig["botToken"]

	// Get the bot id from the bot token
	botId := strings.Split(botToken, ":")[0]
	if botId == "" {
		return model.NewNodeResultWithError(fmt.Errorf("botToken is not valid"))
	}

	if input["tgAuthResult"] != "" {

		// Parse the result
		authResult, err := parseTelegramAuthResult(input["tgAuthResult"], botToken)
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
		if node.CustomConfig["createUser"] == "true" && dbUser == nil {
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
	}

	// Unfortunately with the fragement we cannot know if the user has already logged in or not
	// the the previous node in the history is the telegram login we render the page to load the fragment from the url

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
	origin := getOrigin(state.LoginUri)

	// Assemble the url with query params using net/url
	baseUrl := "https://oauth.telegram.org/auth"
	params := url.Values{}
	params.Set("bot_id", botId)
	params.Set("origin", origin)
	params.Set("return_to", state.LoginUri+"?callback=telegram")

	if node.CustomConfig["requestWriteAccess"] == "true" {
		params.Set("request_access", "write")
	}

	telegramUrl := baseUrl + "?" + params.Encode()

	return model.NewNodeResultWithPrompts(map[string]string{
		"__redirect":   telegramUrl,
		"tgAuthResult": "base64",
	})
}

// parseTelegramAuthResult parses the telegram auth result and verifies the hash
func parseTelegramAuthResult(result string, botToken string) (*TelegramAuthResult, error) {
	return parseTelegramAuthResultWithTime(result, botToken, time.Now())
}

// parseTelegramAuthResultWithTime parses the telegram auth result and verifies the hash
func parseTelegramAuthResultWithTime(result string, botToken string, currentTime time.Time) (*TelegramAuthResult, error) {

	// Base64 decode the result
	decoded, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode telegram auth result: %w", err)
	}

	// Unmarshal the result into a map
	var authResult TelegramAuthResult
	if err := json.Unmarshal(decoded, &authResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal telegram auth result: %w", err)
	}

	// Verify the hash
	if err := verifyTelegramHashWithTime(authResult, botToken, currentTime); err != nil {
		return nil, fmt.Errorf("hash verification failed: %w", err)
	}

	return &authResult, nil
}

// verifyTelegramHashWithTime verifies the hash of the telegram auth result
func verifyTelegramHashWithTime(authResult TelegramAuthResult, botToken string, currentTime time.Time) error {
	// Create a map of all fields except hash for verification
	dataMap := map[string]string{
		"id":         strconv.FormatInt(authResult.ID, 10),
		"first_name": authResult.FirstName,
		"username":   authResult.Username,
		"photo_url":  authResult.PhotoURL,
		"auth_date":  strconv.FormatInt(authResult.AuthDate, 10),
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

	// Compare hashes
	if calculatedHash != authResult.Hash {
		return fmt.Errorf("hash verification failed - data is not from Telegram")
	}

	// Check if data is outdated (more than 15 minutes old)
	currentTimeUnix := currentTime.Unix()
	if (currentTimeUnix - authResult.AuthDate) > TelegramAuthOutdatedThreshold {
		return fmt.Errorf("data is outdated")
	}

	return nil
}

// getOrigin returns the origin of the login uri
func getOrigin(loginUri string) string {

	// parse login uri
	parsedUri, err := url.Parse(loginUri)
	if err != nil {
		return ""
	}

	// Assemble the origin with scheme, host and optionally port
	origin := parsedUri.Scheme + "://" + parsedUri.Host
	return origin
}

type TelegramAuthResult struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}
