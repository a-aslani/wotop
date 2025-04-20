package wotop_logger

import (
	"context"
	"fmt"
	"github.com/Graylog2/go-gelf/gelf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// graylogModel represents a logger model that integrates with Graylog.
//
// Fields:
//   - logger: The underlying zap.Logger instance used for logging.
//   - level: The atomic level configuration for the logger.
//   - graylogAddress: The address of the Graylog server.
//   - stage: The application stage (e.g., development, production).
type graylogModel struct {
	logger         *zap.Logger
	level          zap.AtomicLevel
	graylogAddress string
	stage          string
}

var zapConfig zap.Config

// init initializes the default zap logger configuration.
//
// The configuration specifies JSON encoding, debug level logging, and
// custom encoder settings for time, level, caller, and message formatting.
func init() {
	zapConfig = zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:          "time",
			LevelKey:         "level",
			NameKey:          "logger",
			CallerKey:        "caller",
			MessageKey:       "message",
			LineEnding:       zapcore.DefaultLineEnding,
			EncodeLevel:      zapcore.LowercaseLevelEncoder,
			EncodeTime:       zapcore.ISO8601TimeEncoder,
			EncodeDuration:   zapcore.StringDurationEncoder,
			EncodeCaller:     zapcore.FullCallerEncoder,
			ConsoleSeparator: "\t",
		},
	}
}

// NewGrayLog creates a new instance of graylogModel.
//
// This function initializes a zap logger with a Graylog writer and
// returns the graylogModel instance.
//
// Parameters:
//   - graylogAddress: The address of the Graylog server.
//   - stage: The application stage (e.g., development, production).
//
// Returns:
//   - A pointer to the graylogModel instance.
//   - An error if the Graylog writer could not be created.
func NewGrayLog(graylogAddress string, stage string) (*graylogModel, error) {
	gelfWriter, err := gelf.NewWriter(graylogAddress)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConfig.EncoderConfig),
		zapcore.AddSync(gelfWriter),
		zapConfig.Level,
	)

	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &graylogModel{
		logger:         l,
		graylogAddress: graylogAddress,
	}, nil
}

// Error logs an error message with optional arguments.
//
// Parameters:
//   - ctx: The context for the log entry.
//   - message: The error message to log.
//   - args: Optional arguments to format the message.
func (l *graylogModel) Error(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.logger.Error(messageWithArgs)
}

// Info logs an informational message with optional arguments.
//
// Parameters:
//   - ctx: The context for the log entry.
//   - message: The informational message to log.
//   - args: Optional arguments to format the message.
func (l *graylogModel) Info(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.logger.Info(messageWithArgs)
}

// Warning logs a warning message with optional arguments.
//
// Parameters:
//   - ctx: The context for the log entry.
//   - message: The warning message to log.
//   - args: Optional arguments to format the message.
func (l *graylogModel) Warning(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.logger.Warn(messageWithArgs)
}

// Sync flushes any buffered log entries.
//
// Returns:
//   - An error if the logger could not be synced.
func (l *graylogModel) Sync() error {
	err := l.logger.Sync()
	if err != nil {
		return err
	}
	return nil
}

// log formats a log entry with trace ID and file location information.
//
// Parameters:
//   - ctx: The context containing the trace ID.
//   - data: The data to include in the log entry.
//
// Returns:
//   - A formatted string containing the trace ID, data, and file location.
func (l *graylogModel) log(ctx context.Context, data any) string {
	traceID := GetTraceID(ctx)
	return fmt.Sprintf("%-5s %-60v %s\n", traceID, data, getFileLocationInfo(3))
}
