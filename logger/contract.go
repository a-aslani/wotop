package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

// Logger defines an interface for logging messages at different levels.
//
// Methods:
//   - Info: Logs an informational message.
//   - Error: Logs an error message.
//   - Warning: Logs a warning message.
type Logger interface {
	Info(ctx context.Context, message string, args ...any)
	Error(ctx context.Context, message string, args ...any)
	Warning(ctx context.Context, message string, args ...any)
}

type traceDataType int

const traceDataKey traceDataType = 1 // Key used to store and retrieve trace ID in the context.

// SetTraceID sets a trace ID in the provided context.
//
// Parameters:
//   - ctx: The context in which the trace ID will be set.
//   - traceID: The trace ID to be stored in the context.
//
// Returns:
//   - A new context containing the trace ID.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceDataKey, traceID)
}

// GetTraceID retrieves the trace ID from the provided context.
//
// If no trace ID is found, a default value of "0000000000000000" is returned.
//
// Parameters:
//   - ctx: The context from which the trace ID will be retrieved.
//
// Returns:
//   - A string representing the trace ID.
func GetTraceID(ctx context.Context) string {

	// default traceID
	traceID := "0000000000000000"

	if ctx != nil {
		if v := ctx.Value(traceDataKey); v != nil {
			traceID = v.(string)
		}
	}

	return traceID
}

// getFileLocationInfo retrieves the function's file location information, including
// the filename and line number.
//
// This function uses the runtime package to obtain the caller's information.
// The `skip` parameter determines how many stack frames to ascend.
//
// Parameters:
//   - skip: The number of stack frames to skip when retrieving the caller's information.
//
// Returns:
//   - A string in the format "functionName:lineNumber" or an empty string if the information
//     cannot be retrieved.
func getFileLocationInfo(skip int) string {
	pc, _, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	funcName := runtime.FuncForPC(pc).Name()
	x := strings.LastIndex(funcName, "/")
	return fmt.Sprintf("%s:%d", funcName[x+1:], line)
}

// toJsonString converts an object to its JSON string representation.
//
// This function uses the `json.Marshal` function to serialize the object.
// If an error occurs during marshaling, it is ignored.
//
// Parameters:
//   - obj: The object to be converted to JSON.
//
// Returns:
//   - A string containing the JSON representation of the object.
func toJsonString(obj any) string {
	bytes, _ := json.Marshal(obj)
	return string(bytes)
}
