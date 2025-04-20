package validator

import (
	"context"
	"github.com/a-aslani/wotop/model/apperror"
	"github.com/a-aslani/wotop/model/payload"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// ErrValidationError represents a validation error.
	ErrValidationError apperror.ErrorType = "ER0001 validation error"
	// ErrInvalidTypeInputData indicates an invalid input type.
	ErrInvalidTypeInputData apperror.ErrorType = "ER0002 invalid input type"
	// ErrIsRequired indicates that a required field is missing.
	ErrIsRequired apperror.ErrorType = "ER0003 %s is required"
	// ErrInvalidEmailAddress indicates an invalid email address format.
	ErrInvalidEmailAddress apperror.ErrorType = "ER0004 %s is invalid email address"
	// ErrMaxLen indicates that a field exceeds the maximum allowed length.
	ErrMaxLen apperror.ErrorType = "ER0005 the length of %s must be %d characters or fewer. You entered %d characters"
	// ErrMinLen indicates that a field is below the minimum required length.
	ErrMinLen apperror.ErrorType = "ER0003 the length of %s must be %d characters or longer. You entered %d characters"
)

var (
	// timeType is used to check if a field is of type time.Time.
	timeType = reflect.TypeOf(time.Time{})
)

// Message represents a validation error message.
type Message struct {
	FieldName string `json:"field_name"` // The name of the field that caused the error.
	Code      string `json:"code"`       // The error code.
	Message   string `json:"message"`    // The error message.
}

// validator is a struct that performs validation and stores errors.
type validator struct {
	Errors []any // A list of validation errors.
}

// New creates a new instance of the validator.
//
// Returns:
//   - A pointer to a new validator instance.
func New() *validator {
	return &validator{
		Errors: make([]any, 0),
	}
}

// HttpRequestValidator validates an HTTP request payload.
//
// Parameters:
//   - ctx: The context for managing request-scoped values.
//   - traceID: A unique identifier for tracing the request.
//   - input: The input data to be validated.
//
// Returns:
//   - An error response or nil if validation passes.
//   - An error if validation fails.
func HttpRequestValidator(ctx context.Context, traceID string, input interface{}) (any, error) {

	vld := New()
	isValid, err := vld.Validate(input)
	if err != nil {
		return payload.NewErrorResponse(err, traceID), err
	}

	if !isValid {
		return payload.NewValidationErrorResponse(vld.Errors, traceID), ErrValidationError
	}

	return nil, nil
}

// Validate performs validation on the input data.
//
// Parameters:
//   - input: The input data to be validated.
//
// Returns:
//   - A boolean indicating whether the input is valid.
//   - An error if the input type is invalid.
func (v *validator) Validate(input interface{}) (bool, error) {

	val := reflect.ValueOf(input)

	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct || val.Type().ConvertibleTo(timeType) {
		return false, ErrInvalidTypeInputData
	}

	for i := 0; i < val.NumField(); i++ {

		nameTag := val.Type().Field(i).Tag.Get("name")
		validateTag := val.Type().Field(i).Tag.Get("validate")

		if strings.TrimSpace(validateTag) == "" {
			continue
		}

		name := strings.TrimSpace(nameTag)
		if name == "" {
			name = val.Type().Field(i).Tag.Get("json")
			if name == "" {
				name = val.Type().Field(i).Name
			}
		}

		if err := v.check(name, val.Field(i), validateTag); err != nil {
			return false, err
		}
	}

	return len(v.Errors) == 0, nil
}

// check validates a single field based on its validation rules.
//
// Parameters:
//   - name: The name of the field.
//   - field: The field value to be validated.
//   - validateTag: The validation rules for the field.
//
// Returns:
//   - An error if validation fails.
func (v *validator) check(name string, field reflect.Value, validateTag string) error {

	rules := strings.Split(strings.TrimSpace(validateTag), ",")

	for _, rule := range rules {

		if v.checkHasOldError(name) {
			return nil
		}

		r := strings.Split(strings.TrimSpace(rule), ":")

		switch strings.TrimSpace(r[0]) {
		case "required":
			v.required(name, field)
			break
		case "email":
			v.email(name, field)
			break
		case "min":
			if err := v.min(name, field, r[1]); err != nil {
				return err
			}
			break
		case "max":
			if err := v.max(name, field, r[1]); err != nil {
				return err
			}
			break
		}

	}

	return nil
}

// required checks if a field is non-empty.
//
// Parameters:
//   - name: The name of the field.
//   - field: The field value to be checked.
func (v *validator) required(name string, field reflect.Value) {
	if field.Interface() == reflect.Zero(field.Type()).Interface() {

		err := ErrIsRequired.Var(name)

		v.Errors = append(v.Errors, Message{
			FieldName: name,
			Code:      err.Code(),
			Message:   err.Error(),
		})
	}
}

// email checks if a field contains a valid email address.
//
// Parameters:
//   - name: The name of the field.
//   - field: The field value to be checked.
func (v *validator) email(name string, field reflect.Value) {
	var emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(strings.TrimSpace(field.String())) {

		err := ErrInvalidEmailAddress.Var(strings.TrimSpace(field.String()))

		v.Errors = append(v.Errors, Message{
			FieldName: name,
			Code:      err.Code(),
			Message:   err.Error(),
		})
	}
}

// min checks if a field's length is greater than or equal to a minimum value.
//
// Parameters:
//   - name: The name of the field.
//   - field: The field value to be checked.
//   - params: The minimum length as a string.
//
// Returns:
//   - An error if the field's length is less than the minimum value.
func (v *validator) min(name string, field reflect.Value, params string) error {

	minimum := 1

	var err error

	m := strings.TrimSpace(params)
	if m != "" {
		minimum, err = strconv.Atoi(m)
		if err != nil {
			return err
		}
	}

	if len(strings.TrimSpace(field.String())) < minimum {

		e := ErrMinLen.Var(strings.TrimSpace(name), minimum, len(strings.TrimSpace(field.String())))

		v.Errors = append(v.Errors, Message{
			FieldName: name,
			Code:      e.Code(),
			Message:   e.Error(),
		})
	}

	return nil
}

// max checks if a field's length is less than or equal to a maximum value.
//
// Parameters:
//   - name: The name of the field.
//   - field: The field value to be checked.
//   - params: The maximum length as a string.
//
// Returns:
//   - An error if the field's length exceeds the maximum value.
func (v *validator) max(name string, field reflect.Value, params string) error {

	maximum := 1

	var err error

	m := strings.TrimSpace(params)
	if m != "" {
		maximum, err = strconv.Atoi(m)
		if err != nil {
			return err
		}
	}

	if len(strings.TrimSpace(field.String())) > maximum {

		e := ErrMaxLen.Var(strings.TrimSpace(name), maximum, len(strings.TrimSpace(field.String())))

		v.Errors = append(v.Errors, Message{
			FieldName: name,
			Code:      e.Code(),
			Message:   e.Error(),
		})
	}

	return nil
}

// checkHasOldError checks if a field already has a validation error.
//
// Parameters:
//   - name: The name of the field.
//
// Returns:
//   - A boolean indicating whether the field has a previous error.
func (v *validator) checkHasOldError(name string) bool {
	for _, e := range v.Errors {
		msg := e.(Message)
		if msg.FieldName == name {
			return true
		}
	}
	return false
}
