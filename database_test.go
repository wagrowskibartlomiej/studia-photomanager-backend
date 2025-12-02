package main

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestFindUser(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{File: ":memory:"},
		Photos:   PhotosConfig{Directory: "test_photos"},
		Admin: AdminConfig{
			DefaultLogin:    "testadmin",
			DefaultPassword: "testpass",
		},
		JWT: JWTConfig{SecretKey: "test_key", TimeoutMinutes: 15},
	}

	db, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	user := &User{Login: "testadmin"}
	dbUser, found := FindUser(db, user)
	if !found {
		t.Error("Expected to find default admin user")
	}
	if dbUser.Login != "testadmin" {
		t.Errorf("Expected login testadmin, got %s", dbUser.Login)
	}

	user2 := &User{Login: "nonexistent"}
	_, found2 := FindUser(db, user2)
	if found2 {
		t.Error("Expected not to find nonexistent user")
	}
}

func TestRegisterUser(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{File: ":memory:"},
		Photos:   PhotosConfig{Directory: "test_photos"},
		Admin: AdminConfig{
			DefaultLogin:    "testadmin",
			DefaultPassword: "testpass",
		},
		JWT: JWTConfig{SecretKey: "test_key", TimeoutMinutes: 15},
	}

	db, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	newUser := &User{
		Login:    "newuser",
		Password: "newpass123",
	}

	err = RegisterUser(db, newUser)
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	dbUser, found := FindUser(db, newUser)
	if !found {
		t.Error("Expected to find newly registered user")
	}
	if dbUser.Login != "newuser" {
		t.Errorf("Expected login newuser, got %s", dbUser.Login)
	}
}

func TestLoginUser(t *testing.T) {
	// Valid password
	hash, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if !LoginUser(string(hash), "testpass") {
		t.Error("Expected LoginUser to return true for correct password")
	}

	// Invalid password
	if LoginUser(string(hash), "wrongpass") {
		t.Error("Expected LoginUser to return false for incorrect password")
	}
}

func TestInitDB(t *testing.T) {
	tmpDB, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpDB.Close()
	defer os.Remove(tmpDB.Name())

	cfg := &Config{
		Database: DatabaseConfig{File: tmpDB.Name()},
		Photos:   PhotosConfig{Directory: "test_photos"},
		Admin: AdminConfig{
			DefaultLogin:    "testadmin",
			DefaultPassword: "testpass",
		},
		JWT: JWTConfig{SecretKey: "test_key", TimeoutMinutes: 15},
	}

	db, err := InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// Check table creaton
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query users table: %v", err)
	}

	if count < 1 {
		t.Error("Expected at least one user (default admin)")
	}
}

func TestValidatePassword(t *testing.T) {
	// Easy (default falback) test
	InitPasswordValidator(&Config{
		Password: &PasswordConfig{Mode: "easy"},
	})

	// To short password
	err := ValidatePassword("ab")
	if err == nil {
		t.Error("Expected error for password shorter than 3 characters")
	}

	// Min len password
	err = ValidatePassword("abc")
	if err != nil {
		t.Errorf("Expected no error for password with 3 characters, got: %v", err)
	}

	// No validation
	InitPasswordValidator(&Config{
		Password: &PasswordConfig{Mode: "no-validation"},
	})
	err = ValidatePassword("")
	if err != nil {
		t.Errorf("Expected no error with no-validation mode, got: %v", err)
	}

	// Medium validation
	InitPasswordValidator(&Config{
		Password: &PasswordConfig{Mode: "medium"},
	})
	err = ValidatePassword("abc")
	if err == nil {
		t.Error("Expected error for password without digit in medium mode")
	}
	err = ValidatePassword("abc123")
	if err != nil {
		t.Errorf("Expected no error for valid password in medium mode, got: %v", err)
	}

	// Restrict validation
	InitPasswordValidator(&Config{
		Password: &PasswordConfig{Mode: "restrict"},
	})
	err = ValidatePassword("Password1")
	if err == nil {
		t.Error("Expected error for password without special character in restrict mode")
	}
	err = ValidatePassword("Password1!")
	if err != nil {
		t.Errorf("Expected no error for valid password in restrict mode, got: %v", err)
	}
}

