package entity

import "context"

type Auth interface {
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, sessionID string) error
	IsAuthenticated(ctx context.Context, sessionID string) (bool, error)
}
