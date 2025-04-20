package wotop_validator

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
	ErrValidationError      apperror.ErrorType = "ER0001 validation error"
	ErrInvalidTypeInputData apperror.ErrorType = "ER0002 invalid input type"
	ErrIsRequired           apperror.ErrorType = "ER0003 %s is required"
	ErrInvalidEmailAddress  apperror.ErrorType = "ER0004 %s is invalid email address"
	ErrMaxLen               apperror.ErrorType = "ER0005 the length of %s must be %d characters or fewer. You entered %d characters"
	ErrMinLen               apperror.ErrorType = "ER0003 the length of %s must be %d characters or longer. You entered %d characters"
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

type Message struct {
	FieldName string `json:"field_name"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type validator struct {
	Errors []any
}

func New() *validator {
	return &validator{
		Errors: make([]any, 0),
	}
}

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

func (v *validator) checkHasOldError(name string) bool {
	for _, e := range v.Errors {
		msg := e.(Message)
		if msg.FieldName == name {
			return true
		}
	}
	return false
}
