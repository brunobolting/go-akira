package entity

import "errors"

var ErrNotFound = errors.New("not found")

var ErrUserAlreadyExists = errors.New("user already exists")

var ErrInvalidEmailOrPassword = errors.New("email or password invalid")
