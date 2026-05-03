package logx

import (
	"context"

	"go.uber.org/zap"
)

type noopLogger struct{}

func Noop() Logger {
	return noopLogger{}
}

func (noopLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {}
func (noopLogger) Info(ctx context.Context, msg string, fields ...zap.Field)  {}
func (noopLogger) Warn(ctx context.Context, msg string, fields ...zap.Field)  {}
func (noopLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {}
func (noopLogger) With(fields ...zap.Field) Logger                            { return noopLogger{} }
func (noopLogger) Sync() error                                                { return nil }
