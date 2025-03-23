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
		s.logger.Error(s.ctx, "failed to find user by email", err, map[string]interface{}{"email": email})
		return nil, err
	}
	return u, nil
}
