package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Service string
	Env     string // dev | staging | prod
	Level   zapcore.Level
}

func NewLogger(cfg Config) (*zap.Logger, error) {
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(cfg.Level),
		Development: cfg.Env != "prod",
		Encoding:    "json",

		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",

			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},

		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := zapCfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	logger = logger.With(
		zap.String("service", cfg.Service),
		zap.String("env", cfg.Env),
	)

	return logger, nil
}

type ctxKey struct{}

func WithLogger(ctx context.Context, logger *zap.Logger, requestID string, traceID string) context.Context {
	l := logger.With(
		zap.String("request_id", requestID),
		zap.String("trace_id", traceID),
	)
	return context.WithValue(ctx, ctxKey{}, l)
}

func Logger(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	}
	return zap.L()
}
