package validatorx

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	validatorLib "github.com/go-playground/validator/v10"
)

type Validator interface {
	ValidateStruct(data any) error
}

type validator struct {
	validator *validatorLib.Validate
}

func NewValidator() Validator {
	return &validator{
		validator: validatorLib.New(
			validatorLib.WithRequiredStructEnabled(),
		),
	}
}

func (v *validator) ValidateStruct(input any) error {
	rf := reflect.ValueOf(input)

	if rf.Kind() != reflect.Ptr {
		return fmt.Errorf("[validatorx] input must be a pointer")
	}

	rf = rf.Elem()

	if rf.Kind() != reflect.Struct {
		return fmt.Errorf("[validatorx] input must be a struct")
	}

	if err := v.validateStructRecursive(rf, ""); err != nil {
		return err
	}

	return v.validator.Struct(input)
}

func (v *validator) validateStructRecursive(rf reflect.Value, parentPath string) error {
	structName := rf.Type().Name()
	if parentPath != "" {
		structName = parentPath + "." + structName
	}

	for i := 0; i < rf.NumField(); i++ {
		field := rf.Field(i)
		fieldType := rf.Type().Field(i)

		if !field.CanSet() {
			continue
		}

		validateTag := fieldType.Tag.Get("validate")
		isRequired := strings.Contains(validateTag, "required")

		if field.Type().Kind() == reflect.String {
			oldValue := field.String()
			trimmed := strings.TrimSpace(oldValue)
			field.SetString(trimmed)

			if len(trimmed) != len(oldValue) {
				fieldPath := structName + "." + fieldType.Name
				if parentPath == "" {
					fieldPath = structName + "." + fieldType.Name
				}
				return fmt.Errorf("[validatorx] field '%s' original value not match with sanitized value", fieldPath)
			}

			if isRequired && trimmed == "" {
				fieldPath := structName + "." + fieldType.Name
				return fmt.Errorf("[validatorx] field '%s' is required but empty", fieldPath)
			}
		}

		if field.Type().Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.String {
			if !field.IsNil() {
				stringValue := field.Elem().String()
				trimmed := strings.TrimSpace(stringValue)
				field.Elem().SetString(trimmed)

				if len(trimmed) != len(stringValue) {
					fieldPath := structName + "." + fieldType.Name
					return fmt.Errorf("[validatorx] field '%s' original value not match with sanitized value", fieldPath)
				}

				if isRequired && trimmed == "" {
					fieldPath := structName + "." + fieldType.Name
					return fmt.Errorf("[validatorx] field '%s' is required but empty", fieldPath)
				}
			}
		}

		if shouldSkipType(field.Type()) {
			continue
		}

		if field.Type().Kind() == reflect.Struct {
			currentPath := structName
			if err := v.validateStructRecursive(field, currentPath); err != nil {
				return err
			}
		}

		if field.Type().Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			if !field.IsNil() {
				currentPath := structName
				if err := v.validateStructRecursive(field.Elem(), currentPath); err != nil {
					return err
				}
			}
		}

		if field.Type().Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)

				if item.Kind() == reflect.Struct {
					currentPath := fmt.Sprintf("%s.%s[%d]", structName, fieldType.Name, j)
					if err := v.validateStructRecursive(item, currentPath); err != nil {
						return err
					}
				}

				if item.Kind() == reflect.Ptr && !item.IsNil() && item.Elem().Kind() == reflect.Struct {
					currentPath := fmt.Sprintf("%s.%s[%d]", structName, fieldType.Name, j)
					if err := v.validateStructRecursive(item.Elem(), currentPath); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func shouldSkipType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t == reflect.TypeOf(time.Time{}) {
		return true
	}

	return false
}
