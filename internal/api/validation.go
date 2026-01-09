package api

import (
	"fmt"
	"strings"

	"github.com/intelifox/click-deploy/internal/domain"
)

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (ve *ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ToAppError converts ValidationErrors to AppError
func (ve *ValidationErrors) ToAppError() *domain.AppError {
	return domain.NewValidationError(ve.Error())
}

// ValidateString validates and sanitizes a string field
func ValidateString(value, fieldName string, required bool, minLen, maxLen int) *ValidationErrors {
	errors := &ValidationErrors{}
	
	// Sanitize input first
	value = SanitizeString(value)
	
	if required && strings.TrimSpace(value) == "" {
		errors.Add(fieldName, "is required")
		return errors
	}

	if value != "" {
		if minLen > 0 && len(value) < minLen {
			errors.Add(fieldName, fmt.Sprintf("must be at least %d characters", minLen))
		}
		if maxLen > 0 && len(value) > maxLen {
			errors.Add(fieldName, fmt.Sprintf("must be at most %d characters", maxLen))
		}
	}

	return errors
}

// ValidateInt validates an integer field
func ValidateInt(value *int, fieldName string, required bool, min, max int) *ValidationErrors {
	errors := &ValidationErrors{}

	if required && value == nil {
		errors.Add(fieldName, "is required")
		return errors
	}

	if value != nil {
		if min > 0 && *value < min {
			errors.Add(fieldName, fmt.Sprintf("must be at least %d", min))
		}
		if max > 0 && *value > max {
			errors.Add(fieldName, fmt.Sprintf("must be at most %d", max))
		}
	}

	return errors
}

// ValidateOneOf validates that a value is one of the allowed values
func ValidateOneOf(value, fieldName string, allowedValues []string) *ValidationErrors {
	errors := &ValidationErrors{}

	if value == "" {
		return errors
	}

	valid := false
	for _, allowed := range allowedValues {
		if value == allowed {
			valid = true
			break
		}
	}

	if !valid {
		errors.Add(fieldName, fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", ")))
	}

	return errors
}

// ValidateUUID validates a UUID string
func ValidateUUID(value, fieldName string, required bool) *ValidationErrors {
	errors := &ValidationErrors{}

	if required && value == "" {
		errors.Add(fieldName, "is required")
		return errors
	}

	if value != "" {
		// Basic UUID format validation (8-4-4-4-12 hex digits)
		parts := strings.Split(value, "-")
		if len(parts) != 5 {
			errors.Add(fieldName, "must be a valid UUID")
		}
	}

	return errors
}

// ValidateCreateProjectRequest validates CreateProjectRequest
func ValidateCreateProjectRequest(req *CreateProjectRequest) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validate name
	if nameErrs := ValidateString(req.Name, "name", true, 1, 255); nameErrs.HasErrors() {
		errors.Errors = append(errors.Errors, nameErrs.Errors...)
	}

	// Validate description (optional)
	if req.Description != nil {
		if descErrs := ValidateString(*req.Description, "description", false, 0, 1000); descErrs.HasErrors() {
			errors.Errors = append(errors.Errors, descErrs.Errors...)
		}
	}

	// Validate OpenStack tenant ID (optional)
	if req.OpenStackTenantID != nil {
		if tenantErrs := ValidateString(*req.OpenStackTenantID, "openstack_tenant_id", false, 1, 255); tenantErrs.HasErrors() {
			errors.Errors = append(errors.Errors, tenantErrs.Errors...)
		}
	}

	// Validate default region (optional)
	if req.DefaultRegion != nil {
		if regionErrs := ValidateString(*req.DefaultRegion, "default_region", false, 0, 100); regionErrs.HasErrors() {
			errors.Errors = append(errors.Errors, regionErrs.Errors...)
		}
	}

	return errors
}

// ValidateUpdateProjectRequest validates UpdateProjectRequest
func ValidateUpdateProjectRequest(req *UpdateProjectRequest) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validate name (optional)
	if req.Name != nil {
		if nameErrs := ValidateString(*req.Name, "name", false, 1, 255); nameErrs.HasErrors() {
			errors.Errors = append(errors.Errors, nameErrs.Errors...)
		}
	}

	// Validate description (optional)
	if req.Description != nil {
		if descErrs := ValidateString(*req.Description, "description", false, 0, 1000); descErrs.HasErrors() {
			errors.Errors = append(errors.Errors, descErrs.Errors...)
		}
	}

	// Validate default region (optional)
	if req.DefaultRegion != nil {
		if regionErrs := ValidateString(*req.DefaultRegion, "default_region", false, 0, 100); regionErrs.HasErrors() {
			errors.Errors = append(errors.Errors, regionErrs.Errors...)
		}
	}

	return errors
}

// ValidateCreateServiceRequest validates CreateServiceRequest
func ValidateCreateServiceRequest(req *CreateServiceRequest) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validate name
	if nameErrs := ValidateString(req.Name, "name", true, 1, 255); nameErrs.HasErrors() {
		errors.Errors = append(errors.Errors, nameErrs.Errors...)
	}

	// Validate type
	validTypes := []string{"app", "database", "volume"}
	if typeErrs := ValidateOneOf(req.Type, "type", validTypes); typeErrs.HasErrors() {
		errors.Errors = append(errors.Errors, typeErrs.Errors...)
	}

	// Validate instance size (optional)
	if req.InstanceSize != "" {
		validSizes := []string{"small", "medium", "large", "xlarge"}
		if sizeErrs := ValidateOneOf(req.InstanceSize, "instance_size", validSizes); sizeErrs.HasErrors() {
			errors.Errors = append(errors.Errors, sizeErrs.Errors...)
		}
	}

	// Validate port (optional)
	if portErrs := ValidateInt(req.Port, "port", false, 1, 65535); portErrs.HasErrors() {
		errors.Errors = append(errors.Errors, portErrs.Errors...)
	}

	// Validate git_source_id (optional, must be valid UUID if provided)
	if req.GitSourceID != nil && *req.GitSourceID != "" {
		if uuidErrs := ValidateUUID(*req.GitSourceID, "git_source_id", false); uuidErrs.HasErrors() {
			errors.Errors = append(errors.Errors, uuidErrs.Errors...)
		}
	}

	return errors
}

// ValidateUpdateServiceRequest validates UpdateServiceRequest
func ValidateUpdateServiceRequest(req *UpdateServiceRequest) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validate name (optional)
	if req.Name != nil {
		if nameErrs := ValidateString(*req.Name, "name", false, 1, 255); nameErrs.HasErrors() {
			errors.Errors = append(errors.Errors, nameErrs.Errors...)
		}
	}

	// Validate type (optional)
	if req.Type != nil {
		validTypes := []string{"app", "database", "volume"}
		if typeErrs := ValidateOneOf(*req.Type, "type", validTypes); typeErrs.HasErrors() {
			errors.Errors = append(errors.Errors, typeErrs.Errors...)
		}
	}

	// Validate instance size (optional)
	if req.InstanceSize != nil {
		validSizes := []string{"small", "medium", "large", "xlarge"}
		if sizeErrs := ValidateOneOf(*req.InstanceSize, "instance_size", validSizes); sizeErrs.HasErrors() {
			errors.Errors = append(errors.Errors, sizeErrs.Errors...)
		}
	}

	// Validate port (optional)
	if portErrs := ValidateInt(req.Port, "port", false, 1, 65535); portErrs.HasErrors() {
		errors.Errors = append(errors.Errors, portErrs.Errors...)
	}

	// Validate status (optional)
	if req.Status != nil {
		validStatuses := []string{"pending", "provisioning", "building", "deploying", "live", "failed", "stopped"}
		if statusErrs := ValidateOneOf(*req.Status, "status", validStatuses); statusErrs.HasErrors() {
			errors.Errors = append(errors.Errors, statusErrs.Errors...)
		}
	}

	return errors
}

// ValidateUpdateServicePositionRequest validates UpdateServicePositionRequest
func ValidateUpdateServicePositionRequest(req *UpdateServicePositionRequest) *ValidationErrors {
	errors := &ValidationErrors{}

	// Validate X coordinate
	if req.X < 0 {
		errors.Add("x", "must be non-negative")
	}

	// Validate Y coordinate
	if req.Y < 0 {
		errors.Add("y", "must be non-negative")
	}

	return errors
}

