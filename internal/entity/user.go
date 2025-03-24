package entity

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}

func NewUser(name, email, password string) (*User, error) {
	u := &User{
		ID:        NewID(),
		Name:      name,
		Email:     email,
		Verified:  false,
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u.Password = hash
	return u, nil
}

func (u *User) ComparePassword(password string) bool {
	return ComparePassword(u.Password, password)
}

type UserService interface {
	CreateUser(name, email, password string) (*User, error)
	FindUserByID(id string) (*User, error)
	FindUserByEmail(email string) (*User, error)
}

type UserRepository interface {
	CreateUser(user *User) error
	FindUserByID(id string) (*User, error)
	FindUserByEmail(email string) (*User, error)
}
