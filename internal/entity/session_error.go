package entity

import "errors"

var ErrSessionNotFound = errors.New("session not found")

var ErrSessionExpired = errors.New("session expired")

var ErrUserUnauthorized = errors.New("error.user.unauthorized")

var ErrInvalidSession = errors.New("invalid session")
