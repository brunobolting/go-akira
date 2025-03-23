package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger() *SlogLogger {
	h := slog.NewJSONHandler(os.Stdout, nil)
	return &SlogLogger{
		logger: slog.New(h),
	}
}

func (l *SlogLogger) Info(ctx context.Context, msg string, args map[string]any) {
	if args == nil {
		args = make(map[string]any)
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		args["file"] = file
		args["line"] = line
	}
	attrs := make([]any, 0, len(args)*2)
	for k, v := range args {
		attrs = append(attrs, k, v)
	}
	l.logger.InfoContext(ctx, msg, attrs...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, err error, args map[string]any) {
	if args == nil {
		args = make(map[string]any)
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		args["file"] = file
		args["line"] = line
	}
	attrs := make([]any, 0, len(args)*2+2)
	attrs = append(attrs, "error", err)
	for k, v := range args {
		attrs = append(attrs, k, v)
	}
	l.logger.ErrorContext(ctx, msg, attrs...)
}

func (l *SlogLogger) Close() {}
