package main

import (
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			SecretKey:      "test_secret_key_for_jwt",
			TimeoutMinutes: 15,
		},
	}

	userID := int64(123)
	userLogin := "testuser"

	token, err := GenerateJWT(cfg, userID, userLogin)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	claims, err := parseJWT(cfg, token)
	if err != nil {
		t.Fatalf("Failed to parse generated JWT: %v", err)
	}

	if claims["user_id"].(float64) != float64(userID) {
		t.Errorf("Expected user_id %d, got %f", userID, claims["user_id"].(float64))
	}

	if claims["user_login"].(string) != userLogin {
		t.Errorf("Expected user_login %s, got %s", userLogin, claims["user_login"].(string))
	}

	exp := int64(claims["exp"].(float64))
	now := time.Now().Unix()
	if exp <= now {
		t.Error("Expected expiration time to be in the future")
	}
}

func TestParseJWT(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			SecretKey:      "test_secret_key_for_jwt",
			TimeoutMinutes: 15,
		},
	}

	token, _ := GenerateJWT(cfg, 123, "testuser")
	claims, err := parseJWT(cfg, token)
	if err != nil {
		t.Fatalf("parseJWT failed with valid token: %v", err)
	}

	if claims["user_id"] == nil {
		t.Error("Expected user_id in claims")
	}

	_, err = parseJWT(cfg, "invalid_token")
	if err == nil {
		t.Error("Expected parseJWT to fail with invalid token")
	}

	cfg2 := &Config{
		JWT: JWTConfig{
			SecretKey:      "different_secret_key",
			TimeoutMinutes: 15,
		},
	}
	_, err = parseJWT(cfg2, token)
	if err == nil {
		t.Error("Expected parseJWT to fail with token signed with different key")
	}
}

