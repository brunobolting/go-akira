package session

import (
	"akira/internal/config/env"
	"akira/internal/entity"
	"context"
	"database/sql"
	"net/http"
	"time"
)

func Make(ctx context.Context, db *sql.DB, logger entity.Logger) (entity.SessionService, entity.SessionRepository) {
	repo := NewSessionSqliteRepository(db)
	service := NewService(Options{
		Ctx:      ctx,
		Lifetime: 24 * time.Hour,
		Cookie: &entity.CookieConfig{
			Name:     entity.COOKIE_NAME,
			Path:     "/",
			MaxAge:   86400 * 30 * 6,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
		GCInterval: 5 * time.Minute,
		SecretKey:  env.SESSION_SECRET,
	}, repo, logger)
	return service, repo
}
