package auth

import (
	"akira/internal/entity"
	"context"
	"time"
)

var _ entity.AuthService = (*Service)(nil)

type Service struct {
	user    entity.UserService
	captcha entity.CaptchaService
	logger  entity.Logger
	ctx     context.Context
}

func NewService(
	ctx context.Context,
	user entity.UserService,
	captcha entity.CaptchaService,
	logger entity.Logger,
) *Service {
	return &Service{ctx: ctx, user: user, captcha: captcha, logger: logger}
}

func (s *Service) SignUp(ctx context.Context, req entity.SignUpRequest) (*entity.User, error) {
	if err := req.Validate(); err != nil {
		s.logger.Error(s.ctx, "invalid sign up request", err, map[string]any{
			"name":  req.Name,
			"email": req.Email,
		})
		return nil, err
	}
	if err := s.captcha.Verify(ctx, req.Captcha); err != nil {
		var e entity.RequestError
		return nil, e.Add("captcha", "error.captcha.invalid")
	}
	user, err := s.user.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
		s.logger.Error(s.ctx, "failed to create user", err, map[string]any{"email": req.Email})
		return nil, err
	}
	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	time.Sleep(entity.GetRandomSleep())
	user, err := s.user.FindUserByEmail(email)
	if err != nil || user == nil {
		return nil, entity.ErrInvalidEmailOrPassword
	}
	time.Sleep(entity.GetRandomSleep())
	if !user.ComparePassword(password) {
		return nil, entity.ErrInvalidEmailOrPassword
	}
	return user, nil
}

func (s *Service) IsAuthenticated(ctx context.Context, sessionID string) (bool, error) {
	return false, nil
}
