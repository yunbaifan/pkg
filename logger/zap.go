package logger

import (
	"context"
)

func Debug(ctx context.Context, msg string, fields ...any) {
	Logger.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...any) {
	Logger.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, err error, fields ...any) {
	Logger.Warn(ctx, msg, err, fields...)
}

func Error(ctx context.Context, msg string, err error, fields ...any) {
	Logger.Error(ctx, msg, err, fields...)
}
