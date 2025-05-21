package model

import (
	"encoding/json"
	"testing"
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
