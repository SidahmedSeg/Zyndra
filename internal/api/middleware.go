package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/intelifox/click-deploy/internal/domain"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   domain.ErrorCode `json:"error"`
	Message string            `json:"message"`
	Details string            `json:"details,omitempty"`
}

// ErrorHandler is a middleware that handles errors consistently
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, err error) {
	// Check if it's an AppError
	if appErr, ok := domain.IsAppError(err); ok {
		writeAppError(w, appErr)
		return
	}

	// Check for common database errors
	if err == sql.ErrNoRows {
		writeAppError(w, domain.NewNotFoundError("Resource"))
		return
	}

	// Check for SQL constraint violations
	if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
		writeAppError(w, domain.NewConflictError("Resource already exists"))
		return
	}

	// Default to internal server error
	log.Printf("Unhandled error: %v", err)
	writeAppError(w, domain.ErrInternal.WithError(err))
}

// writeAppError writes an AppError as JSON response
func writeAppError(w http.ResponseWriter, err *domain.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)

	response := ErrorResponse{
		Error:   err.Code,
		Message: err.Message,
	}

	if err.Details != "" {
		response.Details = err.Details
	}

	json.NewEncoder(w).Encode(response)
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteCreated writes a 201 Created response
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

