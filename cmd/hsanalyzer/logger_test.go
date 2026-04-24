package main

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestBuildLoggerSupportsConfiguredLevels(t *testing.T) {
	t.Parallel()

	for _, level := range []string{"debug", "info", "warn", "error"} {
		level := level
		t.Run(level, func(t *testing.T) {
			t.Parallel()

			logger, err := buildLogger(level)
			if err != nil {
				t.Fatalf("build logger: %v", err)
			}
			if logger == nil {
				t.Fatalf("expected logger")
			}
		})
	}
}

func TestBuildLoggerRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := buildLogger("trace")
	if err == nil {
		t.Fatalf("expected invalid level error")
	}
	if !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoggerFromContextFallsBackToNop(t *testing.T) {
	t.Parallel()

	if loggerFromContext(nil) == nil {
		t.Fatalf("expected nop logger for nil context")
	}

	ctx := context.WithValue(context.Background(), loggerContextKey{}, zap.NewNop())
	if loggerFromContext(ctx) == nil {
		t.Fatalf("expected logger from context")
	}
}
