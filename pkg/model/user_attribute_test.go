package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserAttributeWithSocialValue(t *testing.T) {
	// Create a user attribute with a social attribute value
	userAttr := UserAttribute{
		UserID: "123e4567-e89b-12d3-a456-426614174000",
		Tenant: "acme",
		Realm:  "customers",
		Index:  stringPtr("google_1234567890"),
		Type:   "social",
		Value: SocialAttributeValue{
			SocialIDP: "google",
			SocialID:  "1234567890",
		},
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Test JSON encoding
	jsonData, err := json.Marshal(userAttr)
	if err != nil {
		t.Fatalf("Failed to marshal user attribute: %v", err)
	}

	// Test JSON decoding
	var decodedAttr UserAttribute
	err = json.Unmarshal(jsonData, &decodedAttr)
	if err != nil {
		t.Fatalf("Failed to unmarshal user attribute: %v", err)
	}

	// Verify the decoded values match the original
	if decodedAttr.UserID != userAttr.UserID {
		t.Errorf("Expected UserID '%s', got '%s'", userAttr.UserID, decodedAttr.UserID)
	}
	if decodedAttr.Tenant != userAttr.Tenant {
		t.Errorf("Expected Tenant '%s', got '%s'", userAttr.Tenant, decodedAttr.Tenant)
	}
	if decodedAttr.Realm != userAttr.Realm {
		t.Errorf("Expected Realm '%s', got '%s'", userAttr.Realm, decodedAttr.Realm)
	}
	if decodedAttr.Index != userAttr.Index {
		t.Errorf("Expected Index '%s', got '%s'", userAttr.Index, decodedAttr.Index)
	}
	if decodedAttr.Type != userAttr.Type {
		t.Errorf("Expected Type '%s', got '%s'", userAttr.Type, decodedAttr.Type)
	}
	if decodedAttr.CreatedAt != userAttr.CreatedAt {
		t.Errorf("Expected CreatedAt '%v', got '%v'", userAttr.CreatedAt, decodedAttr.CreatedAt)
	}
	if decodedAttr.UpdatedAt != userAttr.UpdatedAt {
		t.Errorf("Expected UpdatedAt '%v', got '%v'", userAttr.UpdatedAt, decodedAttr.UpdatedAt)
	}

	// Verify the social value was properly encoded/decoded
	// Note: The Value field is interface{} so we need to type assert it
	if socialValue, ok := decodedAttr.Value.(map[string]interface{}); ok {
		if socialValue["social_idp"] != "google" {
			t.Errorf("Expected social_idp 'google', got '%v'", socialValue["social_idp"])
		}
		if socialValue["social_id"] != "1234567890" {
			t.Errorf("Expected social_id '1234567890', got '%v'", socialValue["social_id"])
		}
	} else {
		t.Errorf("Failed to type assert Value to map[string]interface{}")
	}
}

func stringPtr(s string) *string {
	return &s
}
