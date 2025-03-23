package entity

import "context"

type Logger interface {
	Info(ctx context.Context, msg string, args map[string]any)
	Error(ctx context.Context, msg string, err error, args map[string]any)
	Close()
}
