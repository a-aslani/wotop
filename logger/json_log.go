package logger

import (
	"context"
	"fmt"
	"github.com/a-aslani/wotop"
	"strings"
	"time"
)

// NewSimpleJSONLogger creates a new instance of a simple JSON logger.
//
// This logger is used to log messages in JSON format with application data and stage information.
//
// Parameters:
//   - appData: The application data containing metadata such as app name and instance ID.
//   - stage: The application stage (e.g., development, production).
//
// Returns:
//   - A Logger instance that logs messages in JSON format.
func NewSimpleJSONLogger(appData wotop.ApplicationData, stage string) Logger {
	return &simpleJSONLoggerImpl{AppData: appData, Stage: stage}
}

// jsonLogModel represents the structure of a JSON log entry.
//
// Fields:
//   - AppName: The name of the application.
//   - AppInstID: The unique instance ID of the application.
//   - Start: The start time of the application.
//   - Severity: The severity level of the log (e.g., INFO, WARNING, ERROR).
//   - Message: The log message.
//   - Location: The location in the code where the log was generated.
//   - Time: The timestamp of the log entry.
type jsonLogModel struct {
	AppName   string `json:"appName"`
	AppInstID string `json:"appInstID"`
	Start     string `json:"start"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Location  string `json:"location"`
	Time      string `json:"time"`
}

// newJSONLogModel creates a new JSON log entry as a string.
//
// This function formats the log entry based on the severity level and includes
// application data, location, and trace ID.
//
// Parameters:
//   - lg: The logger instance containing application data.
//   - flag: The severity level of the log (e.g., INFO, WARNING, ERROR).
//   - loc: The location in the code where the log was generated.
//   - msg: The log message.
//   - trid: The trace ID associated with the log entry.
//
// Returns:
//   - A string representing the JSON log entry.
func newJSONLogModel(lg *simpleJSONLoggerImpl, flag, loc string, msg, trid any) string {

	if flag == "ERROR" {
		return toJsonString(jsonLogModel{
			AppName:   lg.AppData.AppName,
			AppInstID: lg.AppData.AppInstanceID,
			Start:     lg.AppData.StartTime,
			Severity:  flag,
			Message:   fmt.Sprintf("%v %v %v", trid, loc, msg),
			Location:  loc,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	return toJsonString(jsonLogModel{
		AppName:   lg.AppData.AppName,
		AppInstID: lg.AppData.AppInstanceID,
		Start:     lg.AppData.StartTime,
		Severity:  flag,
		Message:   fmt.Sprintf("%v %v", trid, msg),
		Location:  loc,
		Time:      time.Now().Format("2006-01-02 15:04:05"),
	})
}

// simpleJSONLoggerImpl is an implementation of the Logger interface
// that logs messages in JSON format.
//
// Fields:
//   - AppData: The application data containing metadata such as app name and instance ID.
//   - Stage: The application stage (e.g., development, production).
type simpleJSONLoggerImpl struct {
	AppData wotop.ApplicationData
	Stage   string
}

// Warning logs a warning message in JSON format.
//
// This function only logs messages if the application stage is "development".
//
// Parameters:
//   - ctx: The context for the log entry.
//   - message: The warning message to log.
//   - args: Optional arguments to format the message.
func (l simpleJSONLoggerImpl) Warning(ctx context.Context, message string, args ...any) {
	if strings.TrimSpace(strings.ToLower(l.Stage)) != "development" {
		return
	}
	messageWithArgs := fmt.Sprintf(message, args...)
	l.printLog(ctx, "WARNING", messageWithArgs)
}

// Info logs an informational message in JSON format.
//
// This function only logs messages if the application stage is "development".
//
// Parameters:
//   - ctx: The context for the log entry.
//   - message: The informational message to log.
//   - args: Optional arguments to format the message.
func (l simpleJSONLoggerImpl) Info(ctx context.Context, message string, args ...any) {
	if strings.TrimSpace(strings.ToLower(l.Stage)) != "development" {
		return
	}
	messageWithArgs := fmt.Sprintf(message, args...)
	l.printLog(ctx, "INFO", messageWithArgs)
}

// Error logs an error message in JSON format.
//
// This function logs error messages regardless of the application stage.
//
// Parameters:
//   - ctx: The context for the log entry.
//   - message: The error message to log.
//   - args: Optional arguments to format the message.
func (l simpleJSONLoggerImpl) Error(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.printLog(ctx, "ERROR", messageWithArgs)
}

// printLog formats and prints a log entry.
//
// This function includes the trace ID, severity level, and file location
// in the log entry.
//
// Parameters:
//   - ctx: The context containing the trace ID.
//   - flag: The severity level of the log (e.g., INFO, WARNING, ERROR).
//   - data: The log message or data to include in the log entry.
func (l simpleJSONLoggerImpl) printLog(ctx context.Context, flag string, data any) {
	traceID := GetTraceID(ctx)
	fmt.Printf("%-5s %s %-60v %s\n", flag, traceID, data, getFileLocationInfo(3))
	// fmt.Println(newJSONLogModel(&l, flag, getFileLocationInfo(3), data, traceID))
}
