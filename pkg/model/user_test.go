package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserUnmarshalJSON(t *testing.T) {
	testData := `{"id":"","username":"test","display_name":"","email":"","email_verified":false,"phone":"","phone_verified":false,"given_name":"","family_name":"","locale":"","status":"active","roles":[],"groups":[],"attributes":{},"tenant":"acme","realm":"customers","created_at":"","updated_at":"","last_login_at":"","password_locked":false,"webauthn_locked":false,"mfa_locked":false,"failed_login_attempts_password":0,"failed_login_attempts_webauthn":0,"failed_login_attempts_mfa":0,"trusted_devices":[],"federated_id":"","federated_idp":""}`

	var user User
	err := json.Unmarshal([]byte(testData), &user)
	if err != nil {
		t.Fatalf("Failed to unmarshal user: %v", err)
	}

	// Verify the parsed values
	if user.Username != "test" {
		t.Errorf("Expected username 'test', got '%s'", user.Username)
	}
	if user.Tenant != "acme" {
		t.Errorf("Expected tenant 'acme', got '%s'", user.Tenant)
	}
	if user.Realm != "customers" {
		t.Errorf("Expected realm 'customers', got '%s'", user.Realm)
	}
	if user.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", user.Status)
	}
	if !user.CreatedAt.IsZero() {
		t.Errorf("Expected zero CreatedAt, got %v", user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		t.Errorf("Expected zero UpdatedAt, got %v", user.UpdatedAt)
	}
	if user.LastLoginAt != nil {
		t.Errorf("Expected nil LastLoginAt, got %v", user.LastLoginAt)
	}
	if len(user.TrustedDevices) != 0 {
		t.Errorf("Expected empty TrustedDevices, got %v", user.TrustedDevices)
	}
}

func TestUserAttributeHelperMethods(t *testing.T) {
	t.Run("UserWithNoAttributes", func(t *testing.T) {
		user := &User{
			ID:       "user1",
			Tenant:   "acme",
			Realm:    "customers",
			Username: "testuser",
		}

		// Test GetAttributesByType with no attributes
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 0)

		// Test GetAttribute with no attributes
		emailAttr, _, err := GetAttribute[EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.Nil(t, emailAttr)
	})

	t.Run("UserWithValidAttributes", func(t *testing.T) {
		user := &User{
			ID:       "user2",
			Tenant:   "acme",
			Realm:    "customers",
			Username: "testuser2",
			UserAttributes: []UserAttribute{
				{
					ID:        "attr1",
					UserID:    "user2",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "primary@example.com",
					Type:      "email",
					Value:     EmailAttributeValue{Email: "primary@example.com", Verified: true},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "attr2",
					UserID:    "user2",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "+1234567890",
					Type:      "phone",
					Value:     PhoneAttributeValue{Phone: "+1234567890", Verified: false},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// Test GetAttributesByType
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 1)
		assert.Equal(t, "primary@example.com", emailAttrs[0].Index)

		phoneAttrs := user.GetAttributesByType("phone")
		assert.Len(t, phoneAttrs, 1)
		assert.Equal(t, "+1234567890", phoneAttrs[0].Index)

		// Test GetAttribute with single attributes (should work)
		emailAttr, _, err := GetAttribute[EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.NotNil(t, emailAttr)
		assert.Equal(t, "primary@example.com", emailAttr.Email)
		assert.True(t, emailAttr.Verified)

		phoneAttr, _, err := GetAttribute[PhoneAttributeValue](user, "phone")
		assert.NoError(t, err)
		assert.NotNil(t, phoneAttr)
		assert.Equal(t, "+1234567890", phoneAttr.Phone)
		assert.False(t, phoneAttr.Verified)

		// Test GetAttribute with non-existent type
		nonexistentAttr, _, err := GetAttribute[EmailAttributeValue](user, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, nonexistentAttr)
	})

	t.Run("UserWithMultipleAttributesOfSameType", func(t *testing.T) {
		user := &User{
			ID:       "user3",
			Tenant:   "acme",
			Realm:    "customers",
			Username: "testuser3",
			UserAttributes: []UserAttribute{
				{
					ID:        "attr3",
					UserID:    "user3",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "primary@example.com",
					Type:      "email",
					Value:     EmailAttributeValue{Email: "primary@example.com", Verified: true},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "attr4",
					UserID:    "user3",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "work@example.com",
					Type:      "email",
					Value:     EmailAttributeValue{Email: "work@example.com", Verified: false},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// Test GetAttributesByType returns multiple attributes
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 2)
		assert.Equal(t, "primary@example.com", emailAttrs[0].Index)
		assert.Equal(t, "work@example.com", emailAttrs[1].Index)

		// Test GetAttribute with multiple attributes of same type (should return error)
		emailAttr, _, err := GetAttribute[EmailAttributeValue](user, "email")
		assert.Error(t, err)
		assert.Nil(t, emailAttr)
		assert.Contains(t, err.Error(), "multiple attributes of type 'email' found")
		assert.Contains(t, err.Error(), "use GetAttributesByType instead")
	})

	t.Run("UserWithTypeMismatchAttributes", func(t *testing.T) {
		user := &User{
			ID:       "user4",
			Tenant:   "acme",
			Realm:    "customers",
			Username: "testuser4",
			UserAttributes: []UserAttribute{
				{
					ID:        "attr5",
					UserID:    "user4",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "primary@example.com",
					Type:      "email",
					Value:     map[string]interface{}{"email": "primary@example.com", "verified": true},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "attr6",
					UserID:    "user4",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "+1234567890",
					Type:      "phone",
					Value:     "invalid_phone_value", // This should cause a conversion error
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// Test GetAttributesByType
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 1)
		assert.Equal(t, "primary@example.com", emailAttrs[0].Index)

		// Test GetAttribute with map[string]interface{} value (should work via JSON conversion)
		emailAttr, _, err := GetAttribute[EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.NotNil(t, emailAttr)
		assert.Equal(t, "primary@example.com", emailAttr.Email)
		assert.True(t, emailAttr.Verified)

		// Test GetAttribute with invalid value type (should fail)
		phoneAttr, _, err := GetAttribute[PhoneAttributeValue](user, "phone")
		assert.Error(t, err)
		assert.Nil(t, phoneAttr)
		assert.Contains(t, err.Error(), "failed to convert attribute value")
	})

	t.Run("UserWithMixedAttributeTypes", func(t *testing.T) {
		user := &User{
			ID:       "user5",
			Tenant:   "acme",
			Realm:    "customers",
			Username: "testuser5",
			UserAttributes: []UserAttribute{
				{
					ID:        "attr7",
					UserID:    "user5",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "google",
					Type:      "social",
					Value:     SocialAttributeValue{SocialIDP: "google", SocialID: "123456789"},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "attr8",
					UserID:    "user5",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "secret123",
					Type:      "totp",
					Value:     TOTPAttributeValue{SecretKey: "secret123", Locked: false, FailedAttempts: 0},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// Test various attribute types
		socialAttrs := user.GetAttributesByType("social")
		assert.Len(t, socialAttrs, 1)
		assert.Equal(t, "google", socialAttrs[0].Index)

		totpAttrs := user.GetAttributesByType("totp")
		assert.Len(t, totpAttrs, 1)
		assert.Equal(t, "secret123", totpAttrs[0].Index)

		// Test GetAttribute with different types
		socialAttr, _, err := GetAttribute[SocialAttributeValue](user, "social")
		assert.NoError(t, err)
		assert.NotNil(t, socialAttr)
		assert.Equal(t, "google", socialAttr.SocialIDP)
		assert.Equal(t, "123456789", socialAttr.SocialID)

		totpAttr, _, err := GetAttribute[TOTPAttributeValue](user, "totp")
		assert.NoError(t, err)
		assert.NotNil(t, totpAttr)
		assert.Equal(t, "secret123", totpAttr.SecretKey)
		assert.False(t, totpAttr.Locked)
		assert.Equal(t, 0, totpAttr.FailedAttempts)
	})

	t.Run("UserWithMultipleAttributesOfSameType", func(t *testing.T) {
		user := &User{
			ID:       "user6",
			Tenant:   "acme",
			Realm:    "customers",
			Username: "testuser6",
			UserAttributes: []UserAttribute{
				{
					ID:        "attr9",
					UserID:    "user6",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "primary@example.com",
					Type:      "email",
					Value:     EmailAttributeValue{Email: "primary@example.com", Verified: true},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "attr10",
					UserID:    "user6",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "work@example.com",
					Type:      "email",
					Value:     EmailAttributeValue{Email: "work@example.com", Verified: false},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "attr11",
					UserID:    "user6",
					Tenant:    "acme",
					Realm:     "customers",
					Index:     "personal@example.com",
					Type:      "email",
					Value:     EmailAttributeValue{Email: "personal@example.com", Verified: true},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// Test GetAttributesByType returns multiple attributes
		emailAttrs := user.GetAttributesByType("email")
		assert.Len(t, emailAttrs, 3)
		assert.Equal(t, "primary@example.com", emailAttrs[0].Index)
		assert.Equal(t, "work@example.com", emailAttrs[1].Index)
		assert.Equal(t, "personal@example.com", emailAttrs[2].Index)

		// Test GetAttributes with multiple attributes of same type (should work)
		emailValues, _, err := GetAttributes[EmailAttributeValue](user, "email")
		assert.NoError(t, err)
		assert.Len(t, emailValues, 3)

		// Verify the converted values
		assert.Equal(t, "primary@example.com", emailValues[0].Email)
		assert.True(t, emailValues[0].Verified)

		assert.Equal(t, "work@example.com", emailValues[1].Email)
		assert.False(t, emailValues[1].Verified)

		assert.Equal(t, "personal@example.com", emailValues[2].Email)
		assert.True(t, emailValues[2].Verified)

		// Test GetAttribute with multiple attributes of same type (should return error)
		emailAttr, _, err := GetAttribute[EmailAttributeValue](user, "email")
		assert.Error(t, err)
		assert.Nil(t, emailAttr)
		assert.Contains(t, err.Error(), "multiple attributes of type 'email' found")
		assert.Contains(t, err.Error(), "use GetAttributesByType instead")
	})
}
