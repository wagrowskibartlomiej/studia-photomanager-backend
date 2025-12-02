package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type PasswordValidator interface {
	Validate(password string) error
	GetName() string
}

// No validation
type NoValidationValidator struct{}

func (v *NoValidationValidator) Validate(password string) error {
	return nil
}

func (v *NoValidationValidator) GetName() string {
	return "no-validation"
}

// EasyValidator - min len 3
type EasyValidator struct{}

func (v *EasyValidator) Validate(password string) error {
	if len(password) < 3 {
		return fmt.Errorf("password must be at least 3 characters long")
	}
	return nil
}

func (v *EasyValidator) GetName() string {
	return "easy"
}

// MediumValidator - min len 6, at least one letter and one number
type MediumValidator struct{}

func (v *MediumValidator) Validate(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

func (v *MediumValidator) GetName() string {
	return "medium"
}

// RestrictValidator - min len 8, at least one big letter, small letter, number and special sign
type RestrictValidator struct{}

func (v *RestrictValidator) Validate(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) {
			hasSpecial = true
		}
	}

	var errors []string

	if !hasUpper {
		errors = append(errors, "at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "at least one digit")
	}
	if !hasSpecial {
		errors = append(errors, "at least one special character")
	}

	if len(errors) > 0 {
		return fmt.Errorf("password must contain: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (v *RestrictValidator) GetName() string {
	return "restrict"
}

type CustomValidator struct {
	MinLength    int    `json:"min_length"`
	MaxLength    int    `json:"max_length"`
	RequireUpper bool   `json:"require_upper"`
	RequireLower bool   `json:"require_lower"`
	RequireDigit bool   `json:"require_digit"`
	RequireSpecial bool `json:"require_special"`
	Regex        string `json:"regex"`
	regexCompiled *regexp.Regexp
}

func (v *CustomValidator) Validate(password string) error {
	if v.MinLength > 0 && len(password) < v.MinLength {
		return fmt.Errorf("password must be at least %d characters long", v.MinLength)
	}

	if v.MaxLength > 0 && len(password) > v.MaxLength {
		return fmt.Errorf("password must be at most %d characters long", v.MaxLength)
	}

	var errors []string

	if v.RequireUpper {
		hasUpper := false
		for _, char := range password {
			if unicode.IsUpper(char) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			errors = append(errors, "at least one uppercase letter")
		}
	}

	if v.RequireLower {
		hasLower := false
		for _, char := range password {
			if unicode.IsLower(char) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			errors = append(errors, "at least one lowercase letter")
		}
	}

	if v.RequireDigit {
		hasDigit := false
		for _, char := range password {
			if unicode.IsDigit(char) {
				hasDigit = true
				break
			}
		}
		if !hasDigit {
			errors = append(errors, "at least one digit")
		}
	}

	if v.RequireSpecial {
		hasSpecial := false
		for _, char := range password {
			if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			errors = append(errors, "at least one special character")
		}
	}

	if v.Regex != "" {
		if v.regexCompiled == nil {
			var err error
			v.regexCompiled, err = regexp.Compile(v.Regex)
			if err != nil {
				return fmt.Errorf("invalid regex pattern: %v", err)
			}
		}
		if !v.regexCompiled.MatchString(password) {
			errors = append(errors, fmt.Sprintf("password must match pattern: %s", v.Regex))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("password must contain: %s", strings.Join(errors, ", "))
	}

	return nil
}

func (v *CustomValidator) GetName() string {
	return "custom"
}

func GetPasswordValidator(cfg *Config) PasswordValidator {
	if cfg.Password == nil {
		return &NoValidationValidator{}
	}

	switch cfg.Password.Mode {
	case "no-validation":
		return &NoValidationValidator{}
	case "easy":
		return &EasyValidator{}
	case "medium":
		return &MediumValidator{}
	case "restrict":
		return &RestrictValidator{}
	case "custom":
		if cfg.Password.Custom != nil {
			return cfg.Password.Custom
		}
		return &NoValidationValidator{}
	default:
		return &NoValidationValidator{}
	}
}

