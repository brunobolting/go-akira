package user

import (
	"akira/internal/entity"
	"database/sql"
)

func Make(db *sql.DB) (entity.UserService, entity.UserRepository) {
	repo := NewUserSqliteRepository(db)
	service := NewService(repo)
	return service, repo
}
