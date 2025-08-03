package formmap

import (
	"fmt"
	"strings"
)

type ValidationErrors map[string]ValidationField

func (e ValidationErrors) MsgFor(fieldName string) string {
	f, ok := e[fieldName]
	if !ok {
		return ""
	}
	return f.Msg()
}

func (e ValidationErrors) HasError(fieldName string) bool {
	_, ok := e[fieldName]
	return ok
}

type ValidationField struct {
	Tag   string
	Param string
	Field string
}

func (v ValidationField) Msg() string {
	if v.Tag == "" {
		return ""
	}

	switch v.Tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address"
	case "url", "http_url":
		return "This field must be a valid URL"
	case "gte":
		return fmt.Sprintf("Value must be at least %s", v.Param)
	case "lte":
		return fmt.Sprintf("Value must be at most %s", v.Param)
	case "gt":
		return fmt.Sprintf("Value must be greater than %s", v.Param)
	case "lt":
		return fmt.Sprintf("Value must be less than %s", v.Param)
	case "min":
		return fmt.Sprintf("Minimum length is %s", v.Param)
	case "max":
		return fmt.Sprintf("Maximum length is %s", v.Param)
	case "len":
		return fmt.Sprintf("Length must be exactly %s", v.Param)
	case "eq":
		return fmt.Sprintf("Value must be equal to %s", v.Param)
	case "ne":
		return fmt.Sprintf("Value must not be equal to %s", v.Param)
	case "eqfield":
		return fmt.Sprintf("This field must match %s", v.Param)
	case "nefield":
		return fmt.Sprintf("This field must not match %s", v.Param)
	case "not_blank":
		return "This field cannot be empty"
	case "alphanum":
		return "Only alphanumeric characters are allowed"
	case "alpha":
		return "Only alphabetic characters are allowed"
	case "numeric":
		return "Only numeric characters are allowed"
	case "alphanum_with_underscore":
		return "Only alphanumeric characters and underscores are allowed"
	case "mongodb":
		return "Invalid MongoDB ObjectID"
	case "uuid":
		return "Invalid UUID"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", strings.ReplaceAll(v.Param, " ", ", "))
	case "gtcsfield", "gtfield":
		return fmt.Sprintf("Must be greater than %s", v.Param)
	case "ltcsfield", "ltfield":
		return fmt.Sprintf("Must be less than %s", v.Param)
	case "contains":
		return fmt.Sprintf("Must contain '%s'", v.Param)
	case "startswith":
		return fmt.Sprintf("Must start with '%s'", v.Param)
	case "endswith":
		return fmt.Sprintf("Must end with '%s'", v.Param)
	default:
		msg := fmt.Sprintf("Validation failed on '%s' tag", v.Tag)
		if v.Param != "" {
			msg += fmt.Sprintf(" (param: %s)", v.Param)
		}
		return msg
	}
}

func (v ValidationField) String() string {
	return v.Msg()
}

type ValidationError struct {
	Errors ValidationErrors
}

func (v ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validation error"
	}

	var msgs []string
	for field, err := range v.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, err.Msg()))
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

func (v ValidationError) MsgFor(fieldName string) string {
	return v.Errors.MsgFor(fieldName)
}

func (v ValidationError) HasError(fieldName string) bool {
	return v.Errors.HasError(fieldName)
}

func (v ValidationError) IsEmpty() bool {
	return len(v.Errors) == 0
}
