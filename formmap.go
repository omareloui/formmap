package formmap

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type FormInputData struct {
	Value string
	Error string
}

type ValueConverter func(v reflect.Value) string

type FieldMapper func(docField reflect.Value, formField reflect.Value, fieldPath string, valErr ValidationError) error

type Mapper struct {
	converters   map[reflect.Type]ValueConverter
	fieldMappers map[string]FieldMapper
}

func NewMapper() *Mapper {
	m := &Mapper{
		converters:   make(map[reflect.Type]ValueConverter),
		fieldMappers: make(map[string]FieldMapper),
	}

	m.RegisterConverter(reflect.TypeOf(time.Duration(0)), func(v reflect.Value) string {
		d := v.Interface().(time.Duration)
		return strconv.Itoa(int(d.Minutes()))
	})

	m.RegisterConverter(reflect.TypeOf(time.Time{}), func(v reflect.Value) string {
		t := v.Interface().(time.Time)
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339)
	})

	m.RegisterConverter(reflect.TypeOf(float64(0)), func(v reflect.Value) string {
		f := v.Float()
		return strconv.FormatFloat(f, 'f', -1, 64)
	})

	m.RegisterConverter(reflect.TypeOf(float32(0)), func(v reflect.Value) string {
		f := v.Float()
		return strconv.FormatFloat(f, 'f', -1, 32)
	})

	m.RegisterConverter(reflect.TypeOf(int(0)), func(v reflect.Value) string {
		return strconv.Itoa(int(v.Int()))
	})

	m.RegisterConverter(reflect.TypeOf(int64(0)), func(v reflect.Value) string {
		return strconv.FormatInt(v.Int(), 10)
	})

	m.RegisterConverter(reflect.TypeOf(bool(false)), func(v reflect.Value) string {
		return strconv.FormatBool(v.Bool())
	})

	return m
}

func (m *Mapper) RegisterConverter(t reflect.Type, converter ValueConverter) {
	m.converters[t] = converter
}

func (m *Mapper) RegisterFieldMapper(fieldPath string, mapper FieldMapper) {
	m.fieldMappers[fieldPath] = mapper
}

func (m *Mapper) MapToForm(doc any, valErr ValidationError, formData any) error {
	docVal := reflect.ValueOf(doc)
	formVal := reflect.ValueOf(formData)

	if valErr.Errors == nil {
		valErr.Errors = make(ValidationErrors)
	}

	if docVal.Kind() != reflect.Ptr || formVal.Kind() != reflect.Ptr {
		return fmt.Errorf("doc and formData must be pointers")
	}

	if docVal.IsNil() || formVal.IsNil() {
		return fmt.Errorf("doc and formData cannot be nil")
	}

	docVal = docVal.Elem()
	formVal = formVal.Elem()

	return m.mapStruct(docVal, formVal, valErr, "")
}

func (m *Mapper) mapStruct(docVal, formVal reflect.Value, valErr ValidationError, pathPrefix string) error {
	docType := docVal.Type()
	formType := formVal.Type()

	for i := 0; i < docType.NumField(); i++ {
		docField := docType.Field(i)
		docFieldVal := docVal.Field(i)

		if !docField.IsExported() {
			continue
		}

		fieldName := m.getFieldName(docField)
		if fieldName == "-" {
			continue
		}

		formField, found := m.findFormField(formType, fieldName)
		if !found {
			continue
		}

		formFieldVal := formVal.FieldByName(formField.Name)
		if !formFieldVal.IsValid() || !formFieldVal.CanSet() {
			continue
		}

		fieldPath := fieldName
		if pathPrefix != "" {
			fieldPath = pathPrefix + "." + fieldPath
		}

		if mapper, ok := m.fieldMappers[fieldPath]; ok {
			if err := mapper(docFieldVal, formFieldVal, fieldPath, valErr); err != nil {
				return fmt.Errorf("custom mapper for field %s failed: %w", fieldPath, err)
			}
			continue
		}

		if err := m.mapField(docFieldVal, formFieldVal, valErr, fieldPath, formField); err != nil {
			return fmt.Errorf("mapping field %s failed: %w", fieldPath, err)
		}
	}

	return nil
}

func (m *Mapper) getFieldName(field reflect.StructField) string {
	return field.Name
}

func (m *Mapper) findFormField(formType reflect.Type, fieldName string) (reflect.StructField, bool) {
	return formType.FieldByName(fieldName)
}

func (m *Mapper) mapField(docFieldVal, formFieldVal reflect.Value, valErr ValidationError, fieldPath string, formField reflect.StructField) error {
	formFieldType := formField.Type

	if formFieldType.Name() == "FormInputData" {
		return m.mapFormInputData(docFieldVal, formFieldVal, valErr, fieldPath)
	}

	if docFieldVal.Kind() == reflect.Slice && formFieldVal.Kind() == reflect.Slice {
		return m.mapSlice(docFieldVal, formFieldVal, valErr, fieldPath)
	}

	if docFieldVal.Kind() == reflect.Struct && formFieldVal.Kind() == reflect.Struct {
		return m.mapStruct(docFieldVal, formFieldVal, valErr, fieldPath)
	}

	if docFieldVal.Kind() == reflect.Ptr && formFieldVal.Kind() == reflect.Ptr {
		if docFieldVal.IsNil() {
			formFieldVal.Set(reflect.Zero(formFieldVal.Type()))
			return nil
		}

		if formFieldVal.IsNil() {
			formFieldVal.Set(reflect.New(formFieldVal.Type().Elem()))
		}

		return m.mapField(docFieldVal.Elem(), formFieldVal.Elem(), valErr, fieldPath, formField)
	}

	return nil
}

func (m *Mapper) mapFormInputData(docFieldVal, formFieldVal reflect.Value, valErr ValidationError, fieldPath string) error {
	value := m.convertValue(docFieldVal)

	error := valErr.MsgFor(fieldPath)

	valueField := formFieldVal.FieldByName("Value")
	errorField := formFieldVal.FieldByName("Error")

	if valueField.IsValid() && valueField.CanSet() {
		valueField.SetString(value)
	}

	if errorField.IsValid() && errorField.CanSet() {
		errorField.SetString(error)
	}

	return nil
}

func (m *Mapper) mapSlice(docSlice, formSlice reflect.Value, valErr ValidationError, fieldPath string) error {
	if formSlice.Len() != docSlice.Len() {
		newSlice := reflect.MakeSlice(formSlice.Type(), docSlice.Len(), docSlice.Len())

		elemType := formSlice.Type().Elem()
		if elemType.Kind() == reflect.Struct {
			for i := 0; i < newSlice.Len(); i++ {
				newSlice.Index(i).Set(reflect.New(elemType).Elem())
			}
		}

		formSlice.Set(newSlice)
	}

	for i := 0; i < docSlice.Len(); i++ {
		docElem := docSlice.Index(i)
		formElem := formSlice.Index(i)

		indexedPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		if docElem.Kind() == reflect.Struct && formElem.Kind() == reflect.Struct {
			if err := m.mapStruct(docElem, formElem, valErr, indexedPath); err != nil {
				return err
			}
		} else if formElem.Type().Name() == "FormInputData" {
			if err := m.mapFormInputData(docElem, formElem, valErr, indexedPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Mapper) convertValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Bool && v.IsZero() {
		return ""
	}

	if converter, ok := m.converters[v.Type()]; ok {
		return converter(v)
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Interface:
		if !v.IsNil() {
			return m.convertValue(v.Elem())
		}
		return ""
	default:
		return fmt.Sprint(v.Interface())
	}
}

type MapOptions struct {
	FieldConverters map[string]ValueConverter
	SkipFields      []string
}

func (m *Mapper) MapToFormWithOptions(doc any, valErr ValidationError, formData any, opts MapOptions) error {
	for fieldPath, converter := range opts.FieldConverters {
		m.RegisterFieldMapper(fieldPath, func(docField reflect.Value, formField reflect.Value, path string, err ValidationError) error {
			value := converter(docField)
			errorMsg := err.MsgFor(path)

			formField.FieldByName("Value").SetString(value)
			formField.FieldByName("Error").SetString(errorMsg)
			return nil
		})
	}

	return m.MapToForm(doc, valErr, formData)
}
