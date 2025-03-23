package user

import (
	"akira/internal/entity"
	"database/sql"
)

type UserSqliteRepository struct {
	db *sql.DB
}

func NewUserSqliteRepository(db *sql.DB) *UserSqliteRepository {
	return &UserSqliteRepository{db: db}
}

func (r *UserSqliteRepository) scanUserRow(row *sql.Row) (*entity.User, error) {
	var user entity.User
	var nullableAvatar sql.NullString
	err := row.Scan(
		&user.ID,
		&user.Name,
		nullableAvatar,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdateAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	user.Avatar = nullableAvatar.String
	return &user, nil
}

func (r *UserSqliteRepository) FindUserByID(id string) (*entity.User, error) {
	stmt, err := r.db.Prepare("SELECT id, name, avatar, email, password, created_at, update_at FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	return r.scanUserRow(row)
}

func (r *UserSqliteRepository) FindUserByEmail(email string) (*entity.User, error) {
	stmt, err := r.db.Prepare("SELECT id, name, avatar, email, password, created_at, update_at FROM users WHERE email = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(email)
	return r.scanUserRow(row)
}
