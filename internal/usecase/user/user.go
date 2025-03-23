package user

import (
	"akira/internal/entity"
	"context"
	"database/sql"
)

func Make(ctx context.Context, db *sql.DB, logger entity.Logger) (entity.UserService, entity.UserRepository) {
	repo := NewUserSqliteRepository(db)
	service := NewService(ctx, repo, logger)
	return service, repo
}
