package formmap

import (
	"testing"
)

func TestValidationField_Msg(t *testing.T) {
	tests := []struct {
		name     string
		field    ValidationField
		expected string
	}{
		{
			name:     "required tag",
			field:    ValidationField{Tag: "required"},
			expected: "This field is required",
		},
		{
			name:     "email tag",
			field:    ValidationField{Tag: "email"},
			expected: "Invalid email address",
		},
		{
			name:     "min tag with param",
			field:    ValidationField{Tag: "min", Param: "5"},
			expected: "Minimum length is 5",
		},
		{
			name:     "max tag with param",
			field:    ValidationField{Tag: "max", Param: "100"},
			expected: "Maximum length is 100",
		},
		{
			name:     "gte tag with param",
			field:    ValidationField{Tag: "gte", Param: "18"},
			expected: "Value must be at least 18",
		},
		{
			name:     "oneof tag with param",
			field:    ValidationField{Tag: "oneof", Param: "red green blue"},
			expected: "Must be one of: red, green, blue",
		},
		{
			name:     "eqfield tag",
			field:    ValidationField{Tag: "eqfield", Param: "Password"},
			expected: "This field must match Password",
		},
		{
			name:     "url tag",
			field:    ValidationField{Tag: "url"},
			expected: "This field must be a valid URL",
		},
		{
			name:     "alphanum tag",
			field:    ValidationField{Tag: "alphanum"},
			expected: "Only alphanumeric characters are allowed",
		},
		{
			name:     "not_blank tag",
			field:    ValidationField{Tag: "not_blank"},
			expected: "This field cannot be empty",
		},
		{
			name:     "unknown tag without param",
			field:    ValidationField{Tag: "custom_tag"},
			expected: "Validation failed on 'custom_tag' tag",
		},
		{
			name:     "unknown tag with param",
			field:    ValidationField{Tag: "custom_tag", Param: "value"},
			expected: "Validation failed on 'custom_tag' tag (param: value)",
		},
		{
			name:     "empty tag",
			field:    ValidationField{Tag: ""},
			expected: "",
		},
		{
			name:     "contains tag",
			field:    ValidationField{Tag: "contains", Param: "substr"},
			expected: "Must contain 'substr'",
		},
		{
			name:     "startswith tag",
			field:    ValidationField{Tag: "startswith", Param: "prefix"},
			expected: "Must start with 'prefix'",
		},
		{
			name:     "endswith tag",
			field:    ValidationField{Tag: "endswith", Param: "suffix"},
			expected: "Must end with 'suffix'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.Msg()
			if result != tt.expected {
				t.Errorf("Msg() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationField_String(t *testing.T) {
	field := ValidationField{Tag: "required"}
	expected := "This field is required"

	if result := field.String(); result != expected {
		t.Errorf("String() = %v, want %v", result, expected)
	}
}

func TestValidationErrors_MsgFor(t *testing.T) {
	errors := Errors{
		"name":     ValidationField{Tag: "required"},
		"email":    ValidationField{Tag: "email"},
		"age":      ValidationField{Tag: "gte", Param: "18"},
		"password": ValidationField{Tag: "min", Param: "8"},
	}

	tests := []struct {
		fieldName string
		expected  string
	}{
		{"name", "This field is required"},
		{"email", "Invalid email address"},
		{"age", "Value must be at least 18"},
		{"password", "Minimum length is 8"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := errors.MsgFor(tt.fieldName)
			if result != tt.expected {
				t.Errorf("MsgFor(%s) = %v, want %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestValidationErrors_HasError(t *testing.T) {
	errors := Errors{
		"name":  ValidationField{Tag: "required"},
		"email": ValidationField{Tag: "email"},
	}

	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"name", true},
		{"email", true},
		{"age", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := errors.HasError(tt.fieldName)
			if result != tt.expected {
				t.Errorf("HasError(%s) = %v, want %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   Errors
		contains []string
	}{
		{
			name:     "empty errors",
			errors:   Errors{},
			contains: []string{"validation error"},
		},
		{
			name: "single error",
			errors: Errors{
				"name": ValidationField{Tag: "required"},
			},
			contains: []string{"validation failed", "name:", "This field is required"},
		},
		{
			name: "multiple errors",
			errors: Errors{
				"name":  ValidationField{Tag: "required"},
				"email": ValidationField{Tag: "email"},
			},
			contains: []string{"validation failed", "name:", "email:", "This field is required", "Invalid email address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationError{Errors: tt.errors}
			result := err.Error()

			for _, substr := range tt.contains {
				if !contains(result, substr) {
					t.Errorf("Error() = %v, should contain %v", result, substr)
				}
			}
		})
	}
}

func TestValidationError_MsgFor(t *testing.T) {
	valErr := ValidationError{
		Errors: Errors{
			"name": ValidationField{Tag: "required"},
		},
	}

	if msg := valErr.MsgFor("name"); msg != "This field is required" {
		t.Errorf("MsgFor(name) = %v, want 'This field is required'", msg)
	}

	if msg := valErr.MsgFor("nonexistent"); msg != "" {
		t.Errorf("MsgFor(nonexistent) = %v, want empty string", msg)
	}
}

func TestValidationError_HasError(t *testing.T) {
	valErr := ValidationError{
		Errors: Errors{
			"name": ValidationField{Tag: "required"},
		},
	}

	if !valErr.HasError("name") {
		t.Error("HasError(name) should return true")
	}

	if valErr.HasError("email") {
		t.Error("HasError(email) should return false")
	}
}

func TestValidationError_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		valErr   ValidationError
		expected bool
	}{
		{
			name:     "empty validation error",
			valErr:   ValidationError{Errors: Errors{}},
			expected: true,
		},
		{
			name: "validation error with errors",
			valErr: ValidationError{
				Errors: Errors{
					"field": ValidationField{Tag: "required"},
				},
			},
			expected: false,
		},
		{
			name:     "nil errors map",
			valErr:   ValidationError{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.valErr.IsEmpty()
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationField_AllTags(t *testing.T) {
	tags := []string{
		"required", "email", "url", "http_url", "gte", "lte", "gt", "lt",
		"min", "max", "len", "eq", "ne", "eqfield", "nefield",
		"not_blank", "alphanum", "alpha", "numeric", "alphanum_with_underscore",
		"mongodb", "uuid", "oneof", "gtcsfield", "gtfield", "ltcsfield", "ltfield",
		"contains", "startswith", "endswith",
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			field := ValidationField{Tag: tag, Param: "test"}
			msg := field.Msg()
			if msg == "" {
				t.Errorf("Tag %s should return non-empty message", tag)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
