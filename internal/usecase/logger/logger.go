package logger

import (
	"akira/internal/config/env"
	"akira/internal/entity"
	"context"
)

func NewLogger() entity.Logger {
	typ := env.LOGGER_TYPE
	switch typ {
	case "sentry":
		l, err := NewSentryLogger(SentryLoggerOptions{
			Dsn:              env.LOGGER_SENTRY_DSN,
			Environment:      env.ENVIRONMENT,
			TracesSampleRate: env.LOGGER_SENTRY_TRACES_SAMPLE_RATE,
			Debug:            env.LOGGER_SENTRY_DEBUG,
		})
		if err != nil {
			s := NewSlogLogger()
			s.Error(context.Background(), "failed to init sentry logger", err, nil)
			return s
		}
		return l
	default:
		return NewSlogLogger()
	}
}
