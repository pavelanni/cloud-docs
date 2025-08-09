package token

import (
	"strings"
	"testing"
	"time"
)

func TestManager_Generate(t *testing.T) {
	manager := NewManager("test-secret")
	
	token, err := manager.Generate(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	if token == "" {
		t.Error("Generated token is empty")
	}
	
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Errorf("Invalid token format, expected 2 parts, got %d", len(parts))
	}
}

func TestManager_ValidateSuccess(t *testing.T) {
	manager := NewManager("test-secret")
	
	tokenString, err := manager.Generate(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	validToken, err := manager.Validate(tokenString)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if validToken == nil {
		t.Error("Expected valid token, got nil")
	}
	
	if validToken.ID == "" {
		t.Error("Token ID is empty")
	}
	
	if validToken.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired")
	}
}

func TestManager_ValidateExpired(t *testing.T) {
	manager := NewManager("test-secret")
	
	tokenString, err := manager.Generate(-time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	_, err = manager.Validate(tokenString)
	if err == nil {
		t.Error("Expected error for expired token")
	}
	
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected 'expired' error, got: %v", err)
	}
}

func TestManager_ValidateInvalidFormat(t *testing.T) {
	manager := NewManager("test-secret")
	
	tests := []string{
		"",
		"invalid",
		"invalid.token.format",
		"only-one-part",
	}
	
	for _, tokenString := range tests {
		t.Run(tokenString, func(t *testing.T) {
			_, err := manager.Validate(tokenString)
			if err == nil {
				t.Error("Expected error for invalid token format")
			}
		})
	}
}

func TestManager_ValidateInvalidSignature(t *testing.T) {
	manager1 := NewManager("secret1")
	manager2 := NewManager("secret2")
	
	tokenString, err := manager1.Generate(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	_, err = manager2.Validate(tokenString)
	if err == nil {
		t.Error("Expected error for token with invalid signature")
	}
	
	if !strings.Contains(err.Error(), "signature") {
		t.Errorf("Expected 'signature' error, got: %v", err)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"", 24 * time.Hour, false},
		{"24h", 24 * time.Hour, false},
		{"1h30m", 90 * time.Minute, false},
		{"24", 24 * time.Hour, false},
		{"168", 168 * time.Hour, false},
		{"invalid", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}