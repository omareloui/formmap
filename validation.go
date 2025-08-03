package formmap

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(input any) error
	ParseError(err error) ValidationError
}

type PlaygroundValidator struct {
	validator *validator.Validate
}

func NewValidator() *PlaygroundValidator {
	val := validator.New(validator.WithRequiredStructEnabled())

	return &PlaygroundValidator{validator: val}
}

func (v *PlaygroundValidator) Validate(input any) error {
	return v.validator.Struct(input)
}

func (v *PlaygroundValidator) ParseError(err error) ValidationError {
	if err == nil {
		return ValidationError{Errors: ValidationErrors{}}
	}

	valErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return ValidationError{
			Errors: ValidationErrors{
				"_error": ValidationField{
					Tag:   "invalid",
					Field: "_error",
				},
			},
		}
	}

	valerr := ValidationErrors{}
	for _, err := range valErrors {
		namespace := err.Namespace()
		firstDot := strings.Index(namespace, ".")
		path := namespace
		if firstDot > 0 {
			path = namespace[firstDot+1:]
		}

		valerr[path] = ValidationField{
			Tag:   err.ActualTag(),
			Param: err.Param(),
			Field: err.Field(),
		}
	}

	return ValidationError{Errors: valerr}
}

func (v *PlaygroundValidator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}
