package wotop_logger

import (
	"context"
	"fmt"
	"github.com/Graylog2/go-gelf/gelf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type graylogModel struct {
	logger         *zap.Logger
	level          zap.AtomicLevel
	graylogAddress string
	stage          string
}

var zapConfig zap.Config

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

func (l *graylogModel) Error(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.logger.Error(messageWithArgs)
}

func (l *graylogModel) Info(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.logger.Info(messageWithArgs)
}

func (l *graylogModel) Warning(ctx context.Context, message string, args ...any) {
	messageWithArgs := fmt.Sprintf(message, args...)
	l.logger.Warn(messageWithArgs)
}

func (l *graylogModel) Sync() error {
	err := l.logger.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (l *graylogModel) log(ctx context.Context, data any) string {
	traceID := GetTraceID(ctx)
	return fmt.Sprintf("%-5s %-60v %s\n", traceID, data, getFileLocationInfo(3))
}
