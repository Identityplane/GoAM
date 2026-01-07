package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserAttributeWithEmailValue(t *testing.T) {
	// Create a user attribute with an email attribute value
	userAttr := UserAttribute{
		UserID: "123e4567-e89b-12d3-a456-426614174000",
		Tenant: "acme",
		Realm:  "customers",
		Index:  stringPtr("john.doe@example.com"),
		Type:   AttributeTypeEmail,
		Value: EmailAttributeValue{
			Email:    "john.doe@example.com",
			Verified: true,
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
	// Note: UserID, Tenant, and Realm are excluded from JSON (json:"-") so they won't be unmarshaled
	// We'll skip testing these fields since they're not part of the JSON representation

	// Test Index pointer values (not the pointers themselves)
	if (decodedAttr.Index == nil) != (userAttr.Index == nil) {
		t.Errorf("Expected Index nil status to match")
	} else if decodedAttr.Index != nil && userAttr.Index != nil && *decodedAttr.Index != *userAttr.Index {
		t.Errorf("Expected Index '%s', got '%s'", *userAttr.Index, *decodedAttr.Index)
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

	// Verify the email value was properly encoded/decoded
	// Note: The Value field is interface{} so we need to type assert it
	if emailValue, ok := decodedAttr.Value.(map[string]interface{}); ok {
		if emailValue["email"] != "john.doe@example.com" {
			t.Errorf("Expected email 'john.doe@example.com', got '%v'", emailValue["email"])
		}
		if emailValue["verified"] != true {
			t.Errorf("Expected verified 'true', got '%v'", emailValue["verified"])
		}
	} else {
		t.Errorf("Failed to type assert Value to map[string]interface{}")
	}
}

func stringPtr(s string) *string {
	return &s
}
