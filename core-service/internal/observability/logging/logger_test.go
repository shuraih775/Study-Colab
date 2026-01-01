package logging

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger_Builds(t *testing.T) {
	logger, err := NewLogger(Config{
		Service: "test-service",
		Env:     "test",
		Level:   zapcore.InfoLevel,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if logger == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestLoggerFromContext(t *testing.T) {
	base, _ := NewLogger(Config{
		Service: "test",
		Env:     "test",
		Level:   zapcore.InfoLevel,
	})

	ctx := context.Background()
	ctx = WithLogger(ctx, base, "req-1", "trace-1")

	log := Logger(ctx)
	if log == nil {
		t.Fatal("expected logger in context")
	}
}

func TestLogger_IncludesTraceAndRequestID_Improved(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	ctx := WithLogger(context.Background(), logger, "req-123", "trace-456")

	Logger(ctx).Info("user login")

	if recorded.Len() != 1 {
		t.Fatal("expected one log entry")
	}

	entry := recorded.All()[0]

	fields := entry.ContextMap()

	if fields["request_id"] != "req-123" {
		t.Errorf("expected req-123, got %v", fields["request_id"])
	}
	if fields["trace_id"] != "trace-456" {
		t.Errorf("expected trace_id trace-456, got %v", fields["trace_id"])
	}
}
