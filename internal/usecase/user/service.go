package user

import (
	"akira/internal/entity"
	"context"
)

var _ entity.UserService = (*Service)(nil)

type Service struct {
	repo   entity.UserRepository
	logger entity.Logger
	ctx    context.Context
}

func NewService(ctx context.Context, repo entity.UserRepository, logger entity.Logger) *Service {
	return &Service{ctx: ctx, repo: repo, logger: logger}
}

func (s *Service) FindUserByID(id string) (*entity.User, error) {
	u, err := s.repo.FindUserByID(id)
	if err != nil {
		s.logger.Error(s.ctx, "failed to find user by ID", err, map[string]interface{}{"id": id})
		return nil, err
	}
	return u, nil
}

func (s *Service) FindUserByEmail(email string) (*entity.User, error) {
	u, err := s.repo.FindUserByEmail(email)
	if err != nil {
		s.logger.Error(s.ctx, "failed to find user by email", err, map[string]any{"email": email})
		return nil, err
	}
	return u, nil
}

func (s *Service) CreateUser(name, email, password string) (*entity.User, error) {
	exists, err := s.FindUserByEmail(email)
	if err != nil && err != entity.ErrNotFound {
		s.logger.Error(s.ctx, "failed to create user", err, map[string]any{"email": email})
		return nil, err
	}
	if exists != nil {
		return nil, entity.ErrUserAlreadyExists
	}
	u, err := entity.NewUser(name, email, password)
	if err != nil {
		s.logger.Error(s.ctx, "failed to create user", err, map[string]any{"email": email})
		return nil, err
	}
	if err := s.repo.CreateUser(u); err != nil {
		s.logger.Error(s.ctx, "failed to create user", err, map[string]any{"email": email})
		return nil, err
	}
	return u, nil
}
