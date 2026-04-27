package main

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

func buildLogger(level string) (*zap.Logger, error) {
	parsedLevel := zapcore.InfoLevel
	if err := parsedLevel.Set(strings.ToLower(strings.TrimSpace(level))); err != nil {
		return nil, fmt.Errorf("invalid log level %q", level)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(parsedLevel)
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	return cfg.Build()
}

func loggerFromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.NewNop()
	}

	logger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	if !ok || logger == nil {
		return zap.NewNop()
	}

	return logger
}

func syncLogger(logger *zap.Logger) error {
	if logger == nil {
		return nil
	}
	return logger.Sync()
}
