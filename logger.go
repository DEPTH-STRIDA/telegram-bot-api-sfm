package tgfsm

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger creates a new asynchronous JSON logger with configured timestamp formatting,
// log levels and caller information.
func NewZapLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	logger, err := config.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, err
	}
	return logger, nil
}
