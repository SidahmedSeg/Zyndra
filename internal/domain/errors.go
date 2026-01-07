package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"

	// Authentication/Authorization errors
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"

	// Not found errors
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeProjectNotFound ErrorCode = "PROJECT_NOT_FOUND"
	ErrCodeServiceNotFound ErrorCode = "SERVICE_NOT_FOUND"

	// Conflict errors
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	// Internal errors
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase     ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalAPI  ErrorCode = "EXTERNAL_API_ERROR"
)

// AppError represents an application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Err        error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	if e.Details == "" && err != nil {
		e.Details = err.Error()
	}
	return e
}

// Predefined errors
var (
	ErrUnauthorized = NewAppError(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden    = NewAppError(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrNotFound     = NewAppError(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrInternal     = NewAppError(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrDatabase     = NewAppError(ErrCodeDatabase, "Database error", http.StatusInternalServerError)
	ErrValidation   = NewAppError(ErrCodeValidation, "Validation error", http.StatusBadRequest)
)

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string) *AppError {
	return NewAppError(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

