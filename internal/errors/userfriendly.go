package errors

import (
	"fmt"
	"strings"
)

// UserFriendlyError provides user-friendly error messages
type UserFriendlyError struct {
	UserMessage string
	Technical   error
	Code        string
}

func (e *UserFriendlyError) Error() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	if e.Technical != nil {
		return e.Technical.Error()
	}
	return "An unknown error occurred"
}

func (e *UserFriendlyError) Unwrap() error {
	return e.Technical
}

// NewUserFriendlyError creates a new user-friendly error
func NewUserFriendlyError(userMsg string, technical error) *UserFriendlyError {
	return &UserFriendlyError{
		UserMessage: userMsg,
		Technical:   technical,
	}
}

// WithCode adds an error code to the error
func (e *UserFriendlyError) WithCode(code string) *UserFriendlyError {
	e.Code = code
	return e
}

// ToUserMessage converts various error types to user-friendly messages
func ToUserMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's already a user-friendly error
	if ufe, ok := err.(*UserFriendlyError); ok {
		return ufe.UserMessage
	}

	// Check for common error patterns and convert them
	errStr := err.Error()

	// Network errors
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") {
		return "Unable to connect to the service. Please check your network connection and try again."
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return "The operation took too long to complete. Please try again."
	}

	// Database errors
	if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") {
		return "A resource with this name already exists. Please choose a different name."
	}

	if strings.Contains(errStr, "foreign key constraint") {
		return "This resource is being used by another resource and cannot be deleted."
	}

	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist") {
		return "The requested resource was not found. It may have been deleted or you may not have permission to access it."
	}

	// Authentication errors
	if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "authentication") {
		return "You are not authorized to perform this action. Please check your permissions."
	}

	// Validation errors
	if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "validation") {
		return "The provided information is invalid. Please check your input and try again."
	}

	// Infrastructure errors
	if strings.Contains(errStr, "circuit breaker") {
		return "The service is temporarily unavailable. Please try again in a few moments."
	}

	if strings.Contains(errStr, "quota exceeded") || strings.Contains(errStr, "limit exceeded") {
		return "You have reached the resource limit. Please contact support to increase your quota."
	}

	// Build/deployment errors
	if strings.Contains(errStr, "build failed") || strings.Contains(errStr, "buildkit") {
		return "The build process failed. Please check your code and build configuration."
	}

	if strings.Contains(errStr, "deployment failed") {
		return "The deployment failed. Please check the logs for more details."
	}

	// Generic fallback
	return fmt.Sprintf("An error occurred: %s. Please try again or contact support if the problem persists.", errStr)
}

// WrapUserFriendly wraps an error with a user-friendly message
func WrapUserFriendly(userMsg string, err error) error {
	if err == nil {
		return nil
	}
	return NewUserFriendlyError(userMsg, err)
}

