package formmap

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

type TestUser struct {
	ID          string       `json:"id" validate:"required"`
	Name        string       `json:"name" validate:"required,min=3,max=50"`
	Email       string       `json:"email" validate:"required,email"`
	Age         int          `json:"age" validate:"gte=18,lte=120"`
	Password    string       `json:"password" validate:"required,min=8"`
	ConfirmPass string       `json:"confirm_pass" validate:"required,eqfield=Password"`
	Role        string       `json:"role" validate:"required,oneof=admin user guest"`
	Website     string       `json:"website" validate:"omitempty,url"`
	Tags        []string     `json:"tags" validate:"max=5,dive,min=2"`
	Profile     *TestProfile `json:"profile" validate:"omitempty"`
	Settings    TestSettings `json:"settings"`
}

type TestProfile struct {
	Bio      string    `json:"bio" validate:"max=500"`
	Birthday time.Time `json:"birthday" validate:"required"`
}

type TestSettings struct {
	Theme         string `json:"theme" validate:"required,oneof=light dark"`
	Notifications bool   `json:"notifications"`
}

type TestProduct struct {
	Name     string        `json:"name" validate:"required"`
	Variants []TestVariant `json:"variants" validate:"required,min=1,dive"`
}

type TestVariant struct {
	ID    string  `json:"id" validate:"required"`
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"required,gt=0"`
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}

	if v.validator == nil {
		t.Fatal("validator field is nil")
	}
}

func TestPlaygroundValidator_Validate(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name      string
		input     any
		wantError bool
		errorKeys []string
	}{
		{
			name: "valid user",
			input: &TestUser{
				ID:          "user_123",
				Name:        "John Doe",
				Email:       "john@example.com",
				Age:         25,
				Password:    "password123",
				ConfirmPass: "password123",
				Role:        "user",
				Settings: TestSettings{
					Theme: "dark",
				},
			},
			wantError: false,
		},
		{
			name: "invalid user - missing required fields",
			input: &TestUser{
				Email: "not-an-email",
				Age:   17,
				Role:  "superuser",
			},
			wantError: true,
			errorKeys: []string{"ID", "Name", "Email", "Age", "Password", "ConfirmPass", "Role", "Settings.Theme"},
		},
		{
			name: "invalid user - password mismatch",
			input: &TestUser{
				ID:          "user_123",
				Name:        "John",
				Email:       "john@example.com",
				Age:         20,
				Password:    "password123",
				ConfirmPass: "different",
				Role:        "admin",
				Settings: TestSettings{
					Theme: "light",
				},
			},
			wantError: true,
			errorKeys: []string{"ConfirmPass"},
		},
		{
			name: "valid product with variants",
			input: &TestProduct{
				Name: "Test Product",
				Variants: []TestVariant{
					{ID: "v1", Name: "Variant 1", Price: 10.99},
					{ID: "v2", Name: "Variant 2", Price: 20.99},
				},
			},
			wantError: false,
		},
		{
			name: "invalid product - empty variants",
			input: &TestProduct{
				Name:     "Test Product",
				Variants: []TestVariant{},
			},
			wantError: true,
			errorKeys: []string{"Variants"},
		},
		{
			name: "invalid product - invalid variant",
			input: &TestProduct{
				Name: "Test Product",
				Variants: []TestVariant{
					{ID: "v1", Name: "", Price: -5},
				},
			},
			wantError: true,
			errorKeys: []string{"Variants[0].Name", "Variants[0].Price"},
		},
		{
			name:      "nil input",
			input:     nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valErr := v.Validate(tt.input)

			if (valErr != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", valErr, tt.wantError)
				return
			}

			if valErr != nil && len(tt.errorKeys) > 0 {
				for _, key := range tt.errorKeys {
					if !valErr.HasError(key) {
						t.Errorf("Expected error for field %s, but not found", key)
					}
				}
			}
		})
	}
}

func TestPlaygroundValidator_ParseError(t *testing.T) {
	v := NewValidator()

	user := &TestUser{
		Email:    "invalid-email",
		Age:      200,
		Role:     "invalid-role",
		Password: "short",
	}

	valErr := v.Validate(user)
	if valErr == nil {
		t.Fatal("Expected validation error, got nil")
	}

	tests := []struct {
		field    string
		hasError bool
		tag      string
	}{
		{"ID", true, "required"},
		{"Name", true, "required"},
		{"Email", true, "email"},
		{"Age", true, "lte"},
		{"Password", true, "min"},
		{"Role", true, "oneof"},
		{"Settings.Theme", true, "required"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if valErr.HasError(tt.field) != tt.hasError {
				t.Errorf("HasError(%s) = %v, want %v", tt.field, valErr.HasError(tt.field), tt.hasError)
			}

			if tt.hasError {
				fieldErr := valErr.Errors[tt.field]
				if fieldErr.Tag != tt.tag {
					t.Errorf("Error tag for %s = %v, want %v", tt.field, fieldErr.Tag, tt.tag)
				}
			}
		})
	}
}

func TestPlaygroundValidator_ParseError_NilError(t *testing.T) {
	v := NewValidator()

	valErr := v.ParseError(nil)

	if valErr != nil {
		t.Error("ParseError(nil) should return ValidationError with initialized Errors map")
	}
}

func TestPlaygroundValidator_ParseError_NonValidatorError(t *testing.T) {
	v := NewValidator()

	regularErr := &customError{msg: "some error"}

	valErr := v.ParseError(regularErr)

	if !valErr.HasError("_error") {
		t.Error("ParseError(non-validator error) should have '_error' field")
	}

	if valErr.Errors["_error"].Tag != "invalid" {
		t.Error("ParseError(non-validator error) should set tag to 'invalid'")
	}
}

func TestPlaygroundValidator_RegisterValidation(t *testing.T) {
	v := NewValidator()

	err := v.RegisterValidation("even", func(fl validator.FieldLevel) bool {
		num := fl.Field().Int()
		return num%2 == 0
	})
	if err != nil {
		t.Fatalf("RegisterValidation() error = %v", err)
	}

	type testStruct struct {
		Number int `validate:"even"`
	}

	valErr := v.Validate(&testStruct{Number: 4})
	if valErr != nil {
		t.Errorf("Validate() with even number should pass, got error: %v", err)
	}

	valErr = v.Validate(&testStruct{Number: 5})
	if valErr == nil {
		t.Error("Validate() with odd number should fail")
	}
}

func TestPlaygroundValidator_NestedStructs(t *testing.T) {
	v := NewValidator()

	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city" validate:"required"`
	}

	type Person struct {
		Name    string  `json:"name" validate:"required"`
		Address Address `json:"address"`
	}

	person := &Person{
		Name: "John",
		Address: Address{
			City: "New York",
		},
	}

	valErr := v.Validate(person)
	if valErr == nil {
		t.Fatal("Expected validation error")
	}

	if !valErr.HasError("Address.Street") {
		t.Error("Should have error for nested field 'address.street'")
	}
}

func TestPlaygroundValidator_SliceValidation(t *testing.T) {
	v := NewValidator()

	type Item struct {
		Name  string `json:"name" validate:"required"`
		Value int    `json:"value" validate:"gt=0"`
	}

	type Container struct {
		Items []Item `json:"items" validate:"required,min=1,max=3,dive"`
	}

	tests := []struct {
		name      string
		container Container
		wantError bool
		errorKeys []string
	}{
		{
			name: "valid container",
			container: Container{
				Items: []Item{
					{Name: "Item1", Value: 10},
					{Name: "Item2", Value: 20},
				},
			},
			wantError: false,
		},
		{
			name: "empty items",
			container: Container{
				Items: []Item{},
			},
			wantError: true,
			errorKeys: []string{"Items"},
		},
		{
			name: "invalid item in slice",
			container: Container{
				Items: []Item{
					{Name: "Valid", Value: 10},
					{Name: "", Value: -5},
				},
			},
			wantError: true,
			errorKeys: []string{"Items[1].Value"},
		},
		{
			name: "too many items",
			container: Container{
				Items: []Item{
					{Name: "Item1", Value: 1},
					{Name: "Item2", Value: 2},
					{Name: "Item3", Value: 3},
					{Name: "Item4", Value: 4},
				},
			},
			wantError: true,
			errorKeys: []string{"Items"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valErr := v.Validate(&tt.container)

			if (valErr != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", valErr, tt.wantError)
				return
			}

			if valErr != nil && len(tt.errorKeys) > 0 {
				for _, key := range tt.errorKeys {
					if !valErr.HasError(key) {
						t.Errorf("Expected error for field %s, but not found", key)
					}
				}
			}
		})
	}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
