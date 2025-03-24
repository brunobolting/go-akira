package entity

import (
	"akira/internal/config/env"
	"context"
)

type SignUpRequest struct {
	Name     string
	Email    string
	Password string
	Captcha  string
}

func (r SignUpRequest) Validate() RequestError {
	var e RequestError
	if r.Name == "" {
		e = e.Add("name", "error.name.required")
	}
	if len(r.Name) < 3 {
		e = e.Add("name", "error.name.simple.min-length")
	}
	if r.Email == "" {
		e = e.Add("email", "error.email.required")
	}
	if r.Password == "" {
		e = e.Add("password", "error.password.required")
	}
	if len(r.Password) < 8 {
		e = e.Add("password", "error.password.simple.min-length")
	}
	if r.Captcha == "" && env.ISPROD {
		e = e.Add("captcha", "error.captcha.required")
	}
	return e
}

type AuthService interface {
	SignUp(ctx context.Context, req SignUpRequest) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	IsAuthenticated(ctx context.Context, sessionID string) (bool, error)
}

type CaptchaService interface {
	Verify(ctx context.Context, captcha string) error
}
