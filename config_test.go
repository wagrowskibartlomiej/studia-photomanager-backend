package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configContent := `{
		"server": {
			"port": "8080",
			"host": ""
		},
		"database": {
			"file": "./test.db"
		},
		"jwt": {
			"secret_key": "test_secret_key",
			"timeout_minutes": 15
		},
		"photos": {
			"directory": "test_photos"
		},
		"admin": {
			"default_login": "admin",
			"default_password": "admin123"
		}
	}`

	tmpFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Database.File != "./test.db" {
		t.Errorf("Expected database file ./test.db, got %s", cfg.Database.File)
	}

	if cfg.JWT.SecretKey != "test_secret_key" {
		t.Errorf("Expected secret key test_secret_key, got %s", cfg.JWT.SecretKey)
	}

	if cfg.JWT.TimeoutMinutes != 15 {
		t.Errorf("Expected timeout 15 minutes, got %d", cfg.JWT.TimeoutMinutes)
	}

	if cfg.Photos.Directory != "test_photos" {
		t.Errorf("Expected photos directory test_photos, got %s", cfg.Photos.Directory)
	}

	if cfg.Admin.DefaultLogin != "admin" {
		t.Errorf("Expected admin login admin, got %s", cfg.Admin.DefaultLogin)
	}
}

func TestJWTTimeout(t *testing.T) {
	cfg := &JWTConfig{
		TimeoutMinutes: 30,
	}

	timeout := cfg.Timeout()
	if timeout.Minutes() != 30.0 {
		t.Errorf("Expected timeout 30 minutes, got %f", timeout.Minutes())
	}
}

