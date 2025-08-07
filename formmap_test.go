package formmap

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

type TestDocument struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    int
	IsActive    bool
	CreatedAt   time.Time
	Duration    time.Duration
	Tags        []string
	Metadata    TestMetadata
	Items       []TestItem
	OptionalPtr *string
	NestedPtr   *TestMetadata
}

type TestMetadata struct {
	Version string
	Author  string
}

type TestItem struct {
	ItemID   string
	ItemName string
	Price    float64
}

type TestFormData struct {
	ID          FormInputData
	Name        FormInputData
	Description FormInputData
	Price       FormInputData
	Quantity    FormInputData
	IsActive    FormInputData
	CreatedAt   FormInputData
	Duration    FormInputData
	Tags        []FormInputData
	Metadata    TestMetadataForm
	Items       []TestItemForm
	OptionalPtr FormInputData
	NestedPtr   *TestMetadataForm
}

type TestMetadataForm struct {
	Version FormInputData
	Author  FormInputData
}

type TestItemForm struct {
	ItemID   FormInputData
	ItemName FormInputData
	Price    FormInputData
}

func TestNewMapper(t *testing.T) {
	m := NewMapper()

	if m == nil {
		t.Fatal("NewMapper() returned nil")
	}

	if m.converters == nil {
		t.Fatal("converters map is nil")
	}

	if m.fieldMappers == nil {
		t.Fatal("fieldMappers map is nil")
	}
}

func TestMapper_MapToForm_Basic(t *testing.T) {
	mapper := NewMapper()

	doc := &TestDocument{
		ID:          "123",
		Name:        "Test Product",
		Description: "A test product description",
		Price:       99.99,
		Quantity:    10,
		IsActive:    true,
		CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Duration:    90 * time.Minute,
	}

	valErr := ValidationError{
		Errors: Errors{
			"Name":     ValidationField{Tag: "required"},
			"Price":    ValidationField{Tag: "gt", Param: "0"},
			"Quantity": ValidationField{Tag: "gte", Param: "1"},
		},
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, valErr, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	tests := []struct {
		name      string
		gotValue  string
		wantValue string
		gotError  string
		wantError string
	}{
		{
			name:      "ID field",
			gotValue:  formData.ID.Value,
			wantValue: "123",
			gotError:  formData.ID.Error,
			wantError: "",
		},
		{
			name:      "Name field with error",
			gotValue:  formData.Name.Value,
			wantValue: "Test Product",
			gotError:  formData.Name.Error,
			wantError: "This field is required",
		},
		{
			name:      "Price field",
			gotValue:  formData.Price.Value,
			wantValue: "99.99",
			gotError:  formData.Price.Error,
			wantError: "Value must be greater than 0",
		},
		{
			name:      "Quantity field",
			gotValue:  formData.Quantity.Value,
			wantValue: "10",
			gotError:  formData.Quantity.Error,
			wantError: "Value must be at least 1",
		},
		{
			name:      "Boolean field",
			gotValue:  formData.IsActive.Value,
			wantValue: "true",
			gotError:  formData.IsActive.Error,
			wantError: "",
		},
		{
			name:      "Duration field",
			gotValue:  formData.Duration.Value,
			wantValue: "90",
			gotError:  formData.Duration.Error,
			wantError: "",
		},
		{
			name:      "Time field",
			gotValue:  formData.CreatedAt.Value,
			wantValue: "2024-01-01T12:00:00Z",
			gotError:  formData.CreatedAt.Error,
			wantError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gotValue != tt.wantValue {
				t.Errorf("Value = %v, want %v", tt.gotValue, tt.wantValue)
			}
			if tt.gotError != tt.wantError {
				t.Errorf("Error = %v, want %v", tt.gotError, tt.wantError)
			}
		})
	}
}

func TestMapper_MapToForm_NestedStruct(t *testing.T) {
	mapper := NewMapper()

	doc := &TestDocument{
		Metadata: TestMetadata{
			Version: "1.0.0",
			Author:  "John Doe",
		},
	}

	valErr := ValidationError{
		Errors: Errors{
			"Metadata.Version": ValidationField{Tag: "required"},
		},
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, valErr, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if formData.Metadata.Version.Value != "1.0.0" {
		t.Errorf("Nested Version value = %v, want '1.0.0'", formData.Metadata.Version.Value)
	}

	if formData.Metadata.Version.Error != "This field is required" {
		t.Errorf("Nested Version error = %v, want 'This field is required'", formData.Metadata.Version.Error)
	}

	if formData.Metadata.Author.Value != "John Doe" {
		t.Errorf("Nested Author value = %v, want 'John Doe'", formData.Metadata.Author.Value)
	}
}

func TestMapper_MapToForm_Slices(t *testing.T) {
	mapper := NewMapper()

	doc := &TestDocument{
		Tags: []string{"tag1", "tag2", "tag3"},
		Items: []TestItem{
			{ItemID: "item1", ItemName: "First Item", Price: 10.50},
			{ItemID: "item2", ItemName: "Second Item", Price: 20.75},
		},
	}

	valErr := ValidationError{
		Errors: Errors{
			"Items[0].Price":    ValidationField{Tag: "gt", Param: "0"},
			"Items[1].ItemName": ValidationField{Tag: "required"},
		},
	}

	formData := &TestFormData{
		Tags:  make([]FormInputData, len(doc.Tags)),
		Items: make([]TestItemForm, len(doc.Items)),
	}

	err := mapper.MapToForm(doc, valErr, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if len(formData.Tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(formData.Tags))
	}

	for i, tag := range []string{"tag1", "tag2", "tag3"} {
		if formData.Tags[i].Value != tag {
			t.Errorf("Tag[%d] = %v, want %v", i, formData.Tags[i].Value, tag)
		}
	}

	if len(formData.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(formData.Items))
	}

	if formData.Items[0].Price.Value != "10.5" {
		t.Errorf("Items[0].Price value = %v, want '10.5'", formData.Items[0].Price.Value)
	}

	if formData.Items[0].Price.Error != "Value must be greater than 0" {
		t.Errorf("Items[0].Price error = %v, want error message", formData.Items[0].Price.Error)
	}

	if formData.Items[1].ItemName.Error != "This field is required" {
		t.Errorf("Items[1].ItemName error = %v, want 'This field is required'", formData.Items[1].ItemName.Error)
	}
}

func TestMapper_MapToForm_Pointers(t *testing.T) {
	mapper := NewMapper()

	strValue := "optional value"
	doc := &TestDocument{
		OptionalPtr: &strValue,
		NestedPtr: &TestMetadata{
			Version: "2.0.0",
			Author:  "Jane Doe",
		},
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, ValidationError{}, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if formData.OptionalPtr.Value != "optional value" {
		t.Errorf("OptionalPtr value = %v, want 'optional value'", formData.OptionalPtr.Value)
	}

	if formData.NestedPtr == nil {
		t.Fatal("NestedPtr should not be nil")
	}

	if formData.NestedPtr.Version.Value != "2.0.0" {
		t.Errorf("NestedPtr.Version value = %v, want '2.0.0'", formData.NestedPtr.Version.Value)
	}
}

func TestMapper_MapToForm_NilPointers(t *testing.T) {
	mapper := NewMapper()

	doc := &TestDocument{
		OptionalPtr: nil,
		NestedPtr:   nil,
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, ValidationError{}, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if formData.OptionalPtr.Value != "" {
		t.Errorf("OptionalPtr value = %v, want empty string", formData.OptionalPtr.Value)
	}

	if formData.NestedPtr != nil {
		t.Errorf("NestedPtr = %v, want nil", formData.NestedPtr)
	}
}

func TestMapper_RegisterConverter(t *testing.T) {
	mapper := NewMapper()

	mapper.RegisterConverter(reflect.TypeOf(time.Duration(0)), func(v reflect.Value) string {
		d := v.Interface().(time.Duration)
		return d.String()
	})

	doc := &TestDocument{
		Duration: 2*time.Hour + 30*time.Minute,
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, ValidationError{}, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if formData.Duration.Value != "2h30m0s" {
		t.Errorf("Duration value = %v, want '2h30m0s'", formData.Duration.Value)
	}
}

func TestMapper_RegisterFieldMapper(t *testing.T) {
	mapper := NewMapper()

	mapper.RegisterFieldMapper("Price", func(docField, formField reflect.Value, fieldPath string, valErr ValidationError) error {
		price := docField.Float()
		formatted := "$" + strconv.FormatFloat(price, 'f', 2, 64) + " USD"

		formField.FieldByName("Value").SetString(formatted)
		formField.FieldByName("Error").SetString(valErr.MsgFor(fieldPath))
		return nil
	})

	doc := &TestDocument{
		Price: 99.99,
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, ValidationError{}, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if formData.Price.Value != "$99.99 USD" {
		t.Errorf("Price value = %v, want '$99.99 USD'", formData.Price.Value)
	}
}

func TestMapper_MapToFormWithOptions(t *testing.T) {
	mapper := NewMapper()

	doc := &TestDocument{
		Price:    150.00,
		Quantity: 5,
		Items: []TestItem{
			{ItemID: "1", ItemName: "Item 1", Price: 10.00},
			{ItemID: "2", ItemName: "Item 2", Price: 20.00},
		},
	}

	formData := &TestFormData{
		Items: make([]TestItemForm, len(doc.Items)),
	}

	opts := MapOptions{
		FieldConverters: map[string]ValueConverter{
			"Price": func(v reflect.Value) string {
				return "€" + strconv.FormatFloat(v.Float(), 'f', 2, 64)
			},
			"Items[0].Price": func(v reflect.Value) string {
				return "Special: $" + strconv.FormatFloat(v.Float(), 'f', 2, 64)
			},
		},
	}

	err := mapper.MapToFormWithOptions(doc, ValidationError{}, formData, opts)
	if err != nil {
		t.Fatalf("MapToFormWithOptions() error = %v", err)
	}

	if formData.Price.Value != "€150.00" {
		t.Errorf("Price value = %v, want '€150.00'", formData.Price.Value)
	}

	if formData.Items[0].Price.Value != "Special: $10.00" {
		t.Errorf("Items[0].Price value = %v, want 'Special: $10.00'", formData.Items[0].Price.Value)
	}

	if formData.Items[1].Price.Value != "20" {
		t.Errorf("Items[1].Price value = %v, want '20'", formData.Items[1].Price.Value)
	}
}

func TestMapper_ErrorHandling(t *testing.T) {
	mapper := NewMapper()

	tests := []struct {
		name      string
		doc       interface{}
		formData  interface{}
		wantError bool
	}{
		{
			name:      "nil document",
			doc:       nil,
			formData:  &TestFormData{},
			wantError: true,
		},
		{
			name:      "nil form data",
			doc:       &TestDocument{},
			formData:  nil,
			wantError: true,
		},
		{
			name:      "non-pointer document",
			doc:       TestDocument{},
			formData:  &TestFormData{},
			wantError: true,
		},
		{
			name:      "non-pointer form data",
			doc:       &TestDocument{},
			formData:  TestFormData{},
			wantError: true,
		},
		{
			name:      "valid pointers",
			doc:       &TestDocument{},
			formData:  &TestFormData{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapper.MapToForm(tt.doc, ValidationError{}, tt.formData)
			if (err != nil) != tt.wantError {
				t.Errorf("MapToForm() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMapper_ZeroValues(t *testing.T) {
	mapper := NewMapper()

	doc := &TestDocument{
		ID:          "",
		Price:       0,
		Quantity:    0,
		IsActive:    false,
		CreatedAt:   time.Time{},
		Duration:    0,
		Tags:        nil,
		OptionalPtr: nil,
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, ValidationError{}, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	tests := []struct {
		name      string
		gotValue  string
		wantValue string
	}{
		{"empty string", formData.ID.Value, ""},
		{"zero float", formData.Price.Value, ""},
		{"zero int", formData.Quantity.Value, ""},
		{"false bool", formData.IsActive.Value, "false"},
		{"zero time", formData.CreatedAt.Value, ""},
		{"zero duration", formData.Duration.Value, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gotValue != tt.wantValue {
				t.Errorf("Value = %v, want %v", tt.gotValue, tt.wantValue)
			}
		})
	}
}

func TestMapper_CustomTypeConversion(t *testing.T) {
	mapper := NewMapper()

	now := time.Now()
	doc := &TestDocument{
		ID:        "test-id",
		Price:     123.456,
		Quantity:  42,
		IsActive:  true,
		CreatedAt: now,
		Duration:  3*time.Hour + 45*time.Minute,
	}

	formData := &TestFormData{}

	err := mapper.MapToForm(doc, ValidationError{}, formData)
	if err != nil {
		t.Fatalf("MapToForm() error = %v", err)
	}

	if formData.Price.Value != "123.456" {
		t.Errorf("Float conversion = %v, want '123.456'", formData.Price.Value)
	}

	if formData.Quantity.Value != "42" {
		t.Errorf("Int conversion = %v, want '42'", formData.Quantity.Value)
	}

	if formData.IsActive.Value != "true" {
		t.Errorf("Bool conversion = %v, want 'true'", formData.IsActive.Value)
	}

	if formData.Duration.Value != "225" {
		t.Errorf("Duration conversion = %v, want '225'", formData.Duration.Value)
	}

	expectedTime := now.Format(time.RFC3339)
	if formData.CreatedAt.Value != expectedTime {
		t.Errorf("Time conversion = %v, want %v", formData.CreatedAt.Value, expectedTime)
	}
}
