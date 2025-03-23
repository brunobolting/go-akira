package entity

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}

func (u *User) ComparePassword(password string) bool {
	return ComparePassword(u.Password, password)
}

type UserService interface {
	FindUserByID(id string) (*User, error)
	FindUserByEmail(email string) (*User, error)
}

type UserRepository interface {
	FindUserByID(id string) (*User, error)
	FindUserByEmail(email string) (*User, error)
}
