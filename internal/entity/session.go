package entity

import (
	"context"
	"net/http"
	"time"
)

const SESSION_NAME = "akira_session"
const COOKIE_NAME = "akira_cookie"
const REMOTEIP_NAME = "akira_remoteip"

type Session struct {
	ID        string
	UserID    string
	Data      map[string]any
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(ID, userID string, data map[string]any, lifetime time.Duration) *Session {
	now := time.Now().UTC()
	expiresAt := now.Add(lifetime)
	return &Session{
		ID:        ID,
		UserID:    userID,
		Data:      data,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
}

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

type SessionService interface {
	CreateSession(ctx context.Context, userID string) (*Session, error)
	FindSession(ctx context.Context, sessionID string) (*Session, error)
	GetSession(ctx context.Context) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	SetCookie(ctx context.Context, w http.ResponseWriter, sessionID string)
	ClearCookie(ctx context.Context, w http.ResponseWriter)
	GC(ctx context.Context)
	RunGC(ctx context.Context)
	FindExpiredSessions(ctx context.Context) ([]Session, error)
	GenerateSessionID() (string, error)
	SignSessionID(sessionID string) string
	VerifySessionID(signedID string) (string, bool)
	SetSessionMiddleware(next http.Handler) http.Handler
	AuthenticationRequiredMiddleware(w http.ResponseWriter, r *http.Request) error
}

type SessionRepository interface {
	FindSession(id string) (*Session, error)
	CreateSession(session *Session) error
	DeleteSession(id string) error
	GC() error
	GetExpiredSessions() ([]Session, error)
}
