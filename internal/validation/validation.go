package validation

import (
	"fmt"
	"strconv"
	"xui-tg-admin/internal/constants"
)

// ValidateUsername validates a username according to business rules
func ValidateUsername(username string) error {
	if len(username) < constants.MinUsernameLength || len(username) > constants.MaxUsernameLength {
		return fmt.Errorf("username must be between %d and %d characters",
			constants.MinUsernameLength, constants.MaxUsernameLength)
	}

	for _, r := range username {
		if !isValidUsernameChar(r) {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}

	return nil
}

// ValidateDuration validates and parses a duration string
func ValidateDuration(durationStr string) (int, error) {
	days, err := strconv.Atoi(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: must be a number")
	}

	if days < 1 {
		return 0, fmt.Errorf("duration must be at least 1 day")
	}

	if days > 3650 { // 10 years max
		return 0, fmt.Errorf("duration cannot exceed 3650 days")
	}

	return days, nil
}

// isValidUsernameChar checks if a character is valid for usernames
func isValidUsernameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}
