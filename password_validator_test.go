package main

import "testing"

func TestNoValidationValidator(t *testing.T) {
	validator := &NoValidationValidator{}
	
	if err := validator.Validate(""); err != nil {
		t.Errorf("NoValidationValidator should accept empty password, got: %v", err)
	}
	
	if err := validator.Validate("a"); err != nil {
		t.Errorf("NoValidationValidator should accept any password, got: %v", err)
	}
}

func TestEasyValidator(t *testing.T) {
	validator := &EasyValidator{}
	
	if err := validator.Validate("ab"); err == nil {
		t.Error("EasyValidator should reject password shorter than 3 characters")
	}
	
	if err := validator.Validate("abc"); err != nil {
		t.Errorf("EasyValidator should accept password with 3 characters, got: %v", err)
	}
	
	if err := validator.Validate("password123"); err != nil {
		t.Errorf("EasyValidator should accept long password, got: %v", err)
	}
}

func TestMediumValidator(t *testing.T) {
	validator := &MediumValidator{}
	
	// Too short
	if err := validator.Validate("ab"); err == nil {
		t.Error("MediumValidator should reject password shorter than 6 characters")
	}
	
	// Letter missing
	if err := validator.Validate("123456"); err == nil {
		t.Error("MediumValidator should reject password without letter")
	}
	
	// Number missing
	if err := validator.Validate("abcdef"); err == nil {
		t.Error("MediumValidator should reject password without digit")
	}
	
	// Valid
	if err := validator.Validate("abc123"); err != nil {
		t.Errorf("MediumValidator should accept valid password, got: %v", err)
	}
	
	if err := validator.Validate("Password1"); err != nil {
		t.Errorf("MediumValidator should accept valid password, got: %v", err)
	}
}

func TestRestrictValidator(t *testing.T) {
	validator := &RestrictValidator{}
	
	// Too short
	if err := validator.Validate("Pass1!"); err == nil {
		t.Error("RestrictValidator should reject password shorter than 8 characters")
	}
	
	// Big letter missing
	if err := validator.Validate("password1!"); err == nil {
		t.Error("RestrictValidator should reject password without uppercase")
	}
	
	// Small letter missing
	if err := validator.Validate("PASSWORD1!"); err == nil {
		t.Error("RestrictValidator should reject password without lowercase")
	}
	
	// Number missinh
	if err := validator.Validate("Password!"); err == nil {
		t.Error("RestrictValidator should reject password without digit")
	}
	
	// Special Sign missinh
	if err := validator.Validate("Password1"); err == nil {
		t.Error("RestrictValidator should reject password without special character")
	}
	
	// Valid
	if err := validator.Validate("Password1!"); err != nil {
		t.Errorf("RestrictValidator should accept valid password, got: %v", err)
	}
}

func TestCustomValidator(t *testing.T) {
	// Min len test
	validator := &CustomValidator{
		MinLength: 5,
	}
	
	if err := validator.Validate("abcd"); err == nil {
		t.Error("CustomValidator should reject password shorter than min_length")
	}
	
	if err := validator.Validate("abcde"); err != nil {
		t.Errorf("CustomValidator should accept password with min_length, got: %v", err)
	}
	
	// Max len test
	validator2 := &CustomValidator{
		MaxLength: 5,
	}
	
	if err := validator2.Validate("abcdef"); err == nil {
		t.Error("CustomValidator should reject password longer than max_length")
	}
	
	// Requirement test
	validator3 := &CustomValidator{
		MinLength:    6,
		RequireUpper: true,
		RequireLower: true,
		RequireDigit: true,
	}
	
	if err := validator3.Validate("password"); err == nil {
		t.Error("CustomValidator should reject password without uppercase and digit")
	}
	
	if err := validator3.Validate("Password1"); err != nil {
		t.Errorf("CustomValidator should accept valid password, got: %v", err)
	}
	
	// Regex test
	validator4 := &CustomValidator{
		Regex: "^[A-Z][a-z]+[0-9]+$",
	}
	
	if err := validator4.Validate("password1"); err == nil {
		t.Error("CustomValidator should reject password that doesn't match regex")
	}
	
	if err := validator4.Validate("Password1"); err != nil {
		t.Errorf("CustomValidator should accept password matching regex, got: %v", err)
	}
}

func TestGetPasswordValidator(t *testing.T) {
	// Test no-validation
	cfg := &Config{Password: &PasswordConfig{Mode: "no-validation"}}
	validator := GetPasswordValidator(cfg)
	if validator.GetName() != "no-validation" {
		t.Errorf("Expected no-validation, got %s", validator.GetName())
	}
	
	// Test easy
	cfg.Password.Mode = "easy"
	validator = GetPasswordValidator(cfg)
	if validator.GetName() != "easy" {
		t.Errorf("Expected easy, got %s", validator.GetName())
	}
	
	// Test medium
	cfg.Password.Mode = "medium"
	validator = GetPasswordValidator(cfg)
	if validator.GetName() != "medium" {
		t.Errorf("Expected medium, got %s", validator.GetName())
	}
	
	// Test restrict
	cfg.Password.Mode = "restrict"
	validator = GetPasswordValidator(cfg)
	if validator.GetName() != "restrict" {
		t.Errorf("Expected restrict, got %s", validator.GetName())
	}
	
	// Test custom
	cfg.Password.Mode = "custom"
	cfg.Password.Custom = &CustomValidator{MinLength: 5}
	validator = GetPasswordValidator(cfg)
	if validator.GetName() != "custom" {
		t.Errorf("Expected custom, got %s", validator.GetName())
	}
	
	// Test default
	cfg2 := &Config{}
	validator = GetPasswordValidator(cfg2)
	if validator.GetName() != "no-validation" {
		t.Errorf("Expected no-validation as default, got %s", validator.GetName())
	}
}

