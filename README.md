# formmap

[![Go Reference](https://pkg.go.dev/badge/github.com/omareloui/formmap.svg)](https://pkg.go.dev/github.com/omareloui/formmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/omareloui/formmap)](https://goreportcard.com/report/github.com/omareloui/formmap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go package that simplifies form handling by automatically mapping structs to
form data with integrated validation error handling. Perfect for web
applications that need to display validation errors alongside form fields.

## Features

- 🚀 **Automatic Form Mapping** - Maps any struct to form data structures
- ✅ **Integrated Validation** - Built on top of `go-playground/validator`
- 🎯 **Human-Readable Errors** - Converts validation tags to user-friendly messages
- 🔧 **Customizable** - Register custom type converters and field mappers
- 🏗️ **Nested Structure Support** - Handles complex nested structs and slices

## Installation

```bash
go get github.com/omareloui/formmap
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/omareloui/formmap"
)

// Define your domain model
type User struct {
    Name     string `validate:"required,min=3"`
    Email    string `validate:"required,email"`
    Age      int    `validate:"gte=18"`
}

// Define your form structure
type UserForm struct {
    Name  formmap.FormInputData
    Email formmap.FormInputData
    Age   formmap.FormInputData
}

func main() {
    // Create validator and mapper
    validator := formmap.NewValidator()
    mapper := formmap.NewMapper()

    // Your domain object
    user := &User{
        Name:  "Jo",  // Too short
        Email: "bad-email",
        Age:   16,    // Too young
    }

    // Validate
    valErr := validator.Validate(user)

    // Map to form
    form := &UserForm{}
    mapper.MapToForm(user, valErr, form)

    // form.Name.Value = "Jo"
    // form.Name.Error = "Minimum length is 3"
    // form.Email.Error = "Invalid email address"
    // form.Age.Error = "Value must be at least 18"
}
```

## Core Components

### FormInputData

The basic building block for form fields:

```go
type FormInputData struct {
    Value string  // The field value as a string
    Error string  // The validation error message (if any)
}
```

### Validator

Wraps `go-playground/validator` with enhanced error handling:

```go
validator := formmap.NewValidator()

// Validate a struct
valErr := validator.Validate(myStruct)

// Check for specific field errors
if valErr.HasError("email") {
    fmt.Println(valErr.MsgFor("email"))
}
```

### Mapper

Maps structs to form data with automatic type conversion:

```go
mapper := formmap.NewMapper()

// Basic mapping
err := mapper.MapToForm(document, validationError, formData)

// With options
opts := formmap.MapOptions{
    FieldConverters: map[string]formmap.ValueConverter{
        "price": func(v reflect.Value) string {
            return fmt.Sprintf("$%.2f", v.Float())
        },
    },
}
mapper.MapToFormWithOptions(document, validationError, formData, opts)
```

## Advanced Usage

### Custom Type Converters

Register converters for custom types:

```go
// Custom type
type Money struct {
    Amount   float64
    Currency string
}

// Register converter
mapper.RegisterConverter(reflect.TypeOf(Money{}), func(v reflect.Value) string {
    money := v.Interface().(Money)
    return fmt.Sprintf("%.2f %s", money.Amount, money.Currency)
})
```

### Field-Specific Mappers

Override mapping logic for specific fields:

```go
mapper.RegisterFieldMapper("user.created_at", func(docField, formField reflect.Value, path string, valErr *formmap.ValidationError) error {
    t := docField.Interface().(time.Time)
    formatted := t.Format("Jan 2, 2006")

    formField.FieldByName("Value").SetString(formatted)
    formField.FieldByName("Error").SetString(valErr.MsgFor(path))
    return nil
})
```

### Working with Slices

The mapper automatically handles slices and arrays:

```go
type Product struct {
    Name     string
    Variants []Variant `validate:"dive"`
}

type Variant struct {
    SKU   string  `validate:"required"`
    Price float64 `validate:"gt=0"`
}

// Form structures
type ProductForm struct {
    Name     formmap.FormInputData
    Variants []VariantForm
}

type VariantForm struct {
    SKU   formmap.FormInputData
    Price formmap.FormInputData
}

// The mapper handles array indices in validation paths
// e.g., "variants[0].price" -> form.Variants[0].Price.Error
```

### Custom Validation

Register custom validators:

```go
validator.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    // Your phone validation logic
    matched, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, phone)
    return matched
})
```

## Real-World Example

Here's how you might use formmap in an HTTP handler:

```go
func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
    // Parse form data into your domain model
    product := &Product{}
    if err := parseForm(r, product); err != nil {
        http.Error(w, "Bad Request", 400)
        return
    }

    // Validate
    validator := formmap.NewValidator()
    if valErr := validator.Validate(product); err != nil {
        // Map to form with errors
        mapper := formmap.NewMapper()
        form := &ProductForm{}
        mapper.MapToForm(product, valErr, form)

        // Render the form with errors
        renderTemplate(w, "product_form.html", form)
        return
    }

    // Save product...
}
```

## Default Type Conversions

The mapper includes default converters for common types:

- `time.Duration` → minutes as string
- `time.Time` → RFC3339 format
- `float32/float64` → decimal string
- `int/int64/uint` → numeric string
- `bool` → "true" or "false"
- Zero values (except bool) → empty string

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details
