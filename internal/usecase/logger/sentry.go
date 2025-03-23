package logger

import (
	"context"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
)

type SentryLogger struct {
}

type SentryLoggerOptions struct {
	Dsn              string
	Environment      string
	TracesSampleRate float64
	Debug            bool
}

func NewSentryLogger(opts SentryLoggerOptions) (*SentryLogger, error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              opts.Dsn,
		Environment:      opts.Environment,
		TracesSampleRate: opts.TracesSampleRate,
		Debug:            opts.Debug,
	})
	if err != nil {
		return nil, err
	}
	return &SentryLogger{}, nil
}

func (l *SentryLogger) Info(ctx context.Context, msg string, args map[string]any) {
	if args == nil {
		args = make(map[string]any)
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		args["file"] = file
		args["line"] = line
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetContext("args", args)
		sentry.CaptureMessage(msg)
	})
}

func (l *SentryLogger) Error(ctx context.Context, msg string, err error, args map[string]any) {
	if args == nil {
		args = make(map[string]any)
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		args["file"] = file
		args["line"] = line
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetContext("args", args)
		sentry.CaptureException(err)
	})
}

func (l *SentryLogger) Close() {
	sentry.Flush(2 * time.Second)
}
