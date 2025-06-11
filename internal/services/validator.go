package services

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
)

// TextValidator provides text validation functionality
type TextValidator struct {
	logger *logrus.Logger
}

// NewTextValidator creates a new text validator
func NewTextValidator(logger *logrus.Logger) *TextValidator {
	return &TextValidator{
		logger: logger,
	}
}

// ValidateUsername validates a username
func (v *TextValidator) ValidateUsername(username string) error {
	// Check length
	if len(username) < 3 || len(username) > 32 {
		return fmt.Errorf("username must be between 3 and 32 characters")
	}

	// Check if username contains only allowed characters
	if !v.IsLatinOnly(username) {
		return fmt.Errorf("username must contain only Latin letters, numbers, and underscores")
	}

	// Check if username starts with a letter
	if !unicode.IsLetter(rune(username[0])) {
		return fmt.Errorf("username must start with a letter")
	}

	return nil
}

// ParseDuration parses a duration string into a time.Duration
func (v *TextValidator) ParseDuration(text string) (time.Duration, error) {
	// Try to parse as days
	days, err := strconv.Atoi(text)
	if err == nil {
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Try to parse as a standard duration string
	return time.ParseDuration(text)
}

// IsLatinOnly checks if a string contains only Latin letters, numbers, and underscores
func (v *TextValidator) IsLatinOnly(text string) bool {
	for _, r := range text {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
		if unicode.IsLetter(r) && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

// ValidateEmail validates an email address
func (v *TextValidator) ValidateEmail(email string) error {
	// Simple email validation using regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

// ValidatePassword validates a password
func (v *TextValidator) ValidatePassword(password string) error {
	// Check length
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one digit
	hasDigit := false
	for _, r := range password {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

// ValidateServerName validates a server name
func (v *TextValidator) ValidateServerName(name string) error {
	// Check length
	if len(name) < 1 || len(name) > 64 {
		return fmt.Errorf("server name must be between 1 and 64 characters")
	}

	// Check if name contains only allowed characters
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' && r != '.' && r != ' ' {
			return fmt.Errorf("server name contains invalid characters")
		}
	}

	return nil
}

// ValidateURL validates a URL
func (v *TextValidator) ValidateURL(url string) error {
	// Simple URL validation using regex
	urlRegex := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9][-a-zA-Z0-9.]*\.[a-zA-Z]{2,}(:[0-9]+)?(/[-a-zA-Z0-9%_.~#+]*)*(\?[-a-zA-Z0-9%_.~+=&;]*)?$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("invalid URL")
	}
	return nil
}