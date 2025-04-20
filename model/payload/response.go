package payload

import (
	"errors"
	"github.com/a-aslani/wotop.git/model/apperror"
)

type Response struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         any    `json:"data"`
	TraceID      string `json:"trace_id"`
}

func NewSuccessResponse(data any, traceID string) any {
	var res Response
	res.Success = true
	res.Data = data
	res.TraceID = traceID
	return res
}

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
