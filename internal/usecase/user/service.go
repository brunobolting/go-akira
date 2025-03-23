package user

import "akira/internal/entity"

var _ entity.UserService = (*Service)(nil)

type Service struct {
	repo entity.UserRepository
}

func NewService(repo entity.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FindUserByID(id string) (*entity.User, error) {
	return s.repo.FindUserByID(id)
}

func (s *Service) FindUserByEmail(email string) (*entity.User, error) {
	return s.repo.FindUserByEmail(email)
}
