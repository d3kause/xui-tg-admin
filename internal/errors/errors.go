package errors

import (
	"fmt"
)

// ServerNotFoundError represents an error when a server is not found
type ServerNotFoundError struct {
	ServerName string
	UserID     int64
}

// Error returns the error message
func (e *ServerNotFoundError) Error() string {
	return fmt.Sprintf("server not found: %s (user ID: %d)", e.ServerName, e.UserID)
}

// ValidationError represents an error when validation fails
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// XrayAPIError represents an error from the X-ray API
type XrayAPIError struct {
	Operation string
	Status    int
	Message   string
}

// Error returns the error message
func (e *XrayAPIError) Error() string {
	return fmt.Sprintf("X-ray API error during %s (status %d): %s", e.Operation, e.Status, e.Message)
}

// StateError represents an error related to user state
type StateError struct {
	UserID  int64
	State   string
	Message string
}

// Error returns the error message
func (e *StateError) Error() string {
	return fmt.Sprintf("state error for user %d in state %s: %s", e.UserID, e.State, e.Message)
}

// PermissionError represents an error related to permissions
type PermissionError struct {
	UserID         int64
	AccessType     string
	RequiredAccess string
}

// Error returns the error message
func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission error for user %d: has %s access, requires %s access", e.UserID, e.AccessType, e.RequiredAccess)
}

// ConfigError represents an error related to configuration
type ConfigError struct {
	Section string
	Message string
}

// Error returns the error message
func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error in %s: %s", e.Section, e.Message)
}
