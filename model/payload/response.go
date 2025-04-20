package payload

import (
	"errors"
	"github.com/a-aslani/wotop/model/apperror"
)

// Response represents the structure of a standard API response.
//
// Fields:
//   - Success: Indicates whether the operation was successful.
//   - ErrorCode: A code representing the type of error (if any).
//   - ErrorMessage: A message describing the error (if any).
//   - Data: The data payload of the response.
//   - TraceID: A unique identifier for tracing the request.
type Response struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         any    `json:"data"`
	TraceID      string `json:"trace_id"`
}

// NewSuccessResponse creates a new success response.
//
// Parameters:
//   - data: The data payload to include in the response.
//   - traceID: A unique identifier for tracing the request.
//
// Returns:
//   - A Response object with success set to true and the provided data and trace ID.
func NewSuccessResponse(data any, traceID string) any {
	var res Response
	res.Success = true
	res.Data = data
	res.TraceID = traceID
	return res
}

// NewErrorResponse creates a new error response.
//
// Parameters:
//   - err: The error object to include in the response.
//   - traceID: A unique identifier for tracing the request.
//
// Returns:
//   - A Response object with success set to false, the error code, error message, and trace ID.
func NewErrorResponse(err error, traceID string) any {
	var res Response
	res.Success = false
	res.TraceID = traceID

	var et apperror.ErrorType
	ok := errors.As(err, &et)
	if !ok {
		res.ErrorCode = "UNDEFINED"
		res.ErrorMessage = err.Error()
		return res
	}

	res.ErrorCode = et.Code()
	res.ErrorMessage = et.Error()
	return res
}

// NewValidationErrorResponse creates a new validation error response.
//
// Parameters:
//   - messages: A list of validation error messages.
//   - traceID: A unique identifier for tracing the request.
//
// Returns:
//   - A Response object with success set to false, a "BAD_REQUEST" error code,
//     a "validation failed" error message, and the provided validation error messages as data.
func NewValidationErrorResponse(messages []any, traceID string) any {
	var res Response
	res.Success = false
	res.TraceID = traceID

	res.ErrorCode = "BAD_REQUEST"
	res.ErrorMessage = "validation failed"

	res.Data = map[string]any{
		"errors": messages,
	}

	return res
}
