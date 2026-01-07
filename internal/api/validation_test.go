package api

import (
	"testing"
)

func TestValidateString(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		required  bool
		minLen    int
		maxLen    int
		wantError bool
	}{
		{
			name:      "valid string",
			value:     "test",
			fieldName: "name",
			required:  true,
			minLen:    1,
			maxLen:    100,
			wantError: false,
		},
		{
			name:      "required but empty",
			value:     "",
			fieldName: "name",
			required:  true,
			minLen:    1,
			maxLen:    100,
			wantError: true,
		},
		{
			name:      "optional and empty",
			value:     "",
			fieldName: "description",
			required:  false,
			minLen:    1,
			maxLen:    100,
			wantError: false,
		},
		{
			name:      "too short",
			value:     "ab",
			fieldName: "name",
			required:  true,
			minLen:    3,
			maxLen:    100,
			wantError: true,
		},
		{
			name:      "too long",
			value:     string(make([]byte, 101)),
			fieldName: "name",
			required:  true,
			minLen:    1,
			maxLen:    100,
			wantError: true,
		},
		{
			name:      "whitespace only (required)",
			value:     "   ",
			fieldName: "name",
			required:  true,
			minLen:    1,
			maxLen:    100,
			wantError: true,
		},
		{
			name:      "whitespace trimmed",
			value:     "  test  ",
			fieldName: "name",
			required:  true,
			minLen:    1,
			maxLen:    100,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateString(tt.value, tt.fieldName, tt.required, tt.minLen, tt.maxLen)
			hasErrors := errors.HasErrors()

			if hasErrors != tt.wantError {
				t.Errorf("ValidateString() hasErrors = %v, want %v. Errors: %v", hasErrors, tt.wantError, errors.Error())
			}
		})
	}
}

func TestValidateInt(t *testing.T) {
	tests := []struct {
		name      string
		value     *int
		fieldName string
		required  bool
		min       int
		max       int
		wantError bool
	}{
		{
			name:      "valid int",
			value:     intPtr(50),
			fieldName: "port",
			required:  true,
			min:       1,
			max:       65535,
			wantError: false,
		},
		{
			name:      "required but nil",
			value:     nil,
			fieldName: "port",
			required:  true,
			min:       1,
			max:       65535,
			wantError: true,
		},
		{
			name:      "optional and nil",
			value:     nil,
			fieldName: "timeout",
			required:  false,
			min:       1,
			max:       100,
			wantError: false,
		},
		{
			name:      "too small",
			value:     intPtr(0),
			fieldName: "port",
			required:  true,
			min:       1,
			max:       65535,
			wantError: true,
		},
		{
			name:      "too large",
			value:     intPtr(70000),
			fieldName: "port",
			required:  true,
			min:       1,
			max:       65535,
			wantError: true,
		},
		{
			name:      "at minimum",
			value:     intPtr(1),
			fieldName: "port",
			required:  true,
			min:       1,
			max:       65535,
			wantError: false,
		},
		{
			name:      "at maximum",
			value:     intPtr(65535),
			fieldName: "port",
			required:  true,
			min:       1,
			max:       65535,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateInt(tt.value, tt.fieldName, tt.required, tt.min, tt.max)
			hasErrors := errors.HasErrors()

			if hasErrors != tt.wantError {
				t.Errorf("ValidateInt() hasErrors = %v, want %v. Errors: %v", hasErrors, tt.wantError, errors.Error())
			}
		})
	}
}

func TestValidateOneOf(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		fieldName    string
		allowedValues []string
		wantError    bool
	}{
		{
			name:         "valid value",
			value:        "app",
			fieldName:    "type",
			allowedValues: []string{"app", "database", "volume"},
			wantError:    false,
		},
		{
			name:         "invalid value",
			value:        "invalid",
			fieldName:    "type",
			allowedValues: []string{"app", "database", "volume"},
			wantError:    true,
		},
		{
			name:         "empty value (optional)",
			value:        "",
			fieldName:    "type",
			allowedValues: []string{"app", "database", "volume"},
			wantError:    false,
		},
		{
			name:         "case sensitive",
			value:        "App",
			fieldName:    "type",
			allowedValues: []string{"app", "database", "volume"},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateOneOf(tt.value, tt.fieldName, tt.allowedValues)
			hasErrors := errors.HasErrors()

			if hasErrors != tt.wantError {
				t.Errorf("ValidateOneOf() hasErrors = %v, want %v. Errors: %v", hasErrors, tt.wantError, errors.Error())
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	errors := &ValidationErrors{}

	// Test HasErrors
	if errors.HasErrors() {
		t.Error("New ValidationErrors should not have errors")
	}

	// Test Add
	errors.Add("field1", "error 1")
	errors.Add("field2", "error 2")

	if !errors.HasErrors() {
		t.Error("ValidationErrors should have errors after Add")
	}

	if len(errors.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors.Errors))
	}

	// Test Error() method
	errStr := errors.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}

	// Test ToAppError
	appErr := errors.ToAppError()
	if appErr == nil {
		t.Error("ToAppError() should return non-nil AppError")
	}
}

func intPtr(i int) *int {
	return &i
}

