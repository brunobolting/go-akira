package book

import (
	"akira/internal/entity"
	"context"
	"fmt"
)

var _ entity.BookService = (*Service)(nil)

type Service struct {
	ctx    context.Context
	repo   entity.BookRepository
	logger entity.Logger
}

func NewService(ctx context.Context, repo entity.BookRepository, logger entity.Logger) *Service {
	return &Service{
		ctx:    ctx,
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateBook(userID string, req entity.CreateBookRequest) (*entity.Book, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	slug, err := s.ensureUniqueSlug(userID, req.Name)
	if err != nil {
		s.logger.Error(s.ctx, "failed to ensure unique slug", err, map[string]any{
			"user_id": userID,
			"name":    req.Name,
		})
		return nil, err
	}
	book := entity.NewBook(
		userID,
		req.Name,
		req.Edition,
		req.Description,
		slug,
		req.CoverImage,
		req.PageCount,
		req.Volume,
		req.Rating,
		req.Publisher,
		req.Author,
		req.ISBN,
		req.Tags,
		req.Metadata,
		req.Language,
	)
	if err := s.repo.CreateBook(book); err != nil {
		s.logger.Error(s.ctx, "failed to create book", err, map[string]any{
			"user_id": userID,
			"name":    req.Name,
		})
		return nil, err
	}
	return book, nil
}

func (s *Service) CreateCollectionBook(userID, collectionID string, req entity.CreateBookRequest) (*entity.Book, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	slug, err := s.ensureUniqueSlug(userID, req.Name)
	if err != nil {
		s.logger.Error(s.ctx, "failed to ensure unique slug", err, map[string]any{
			"user_id": userID,
			"name":    req.Name,
		})
		return nil, err
	}
	book := entity.NewBook(
		userID,
		req.Name,
		req.Edition,
		req.Description,
		slug,
		req.CoverImage,
		req.PageCount,
		req.Volume,
		req.Rating,
		req.Publisher,
		req.Author,
		req.ISBN,
		req.Tags,
		req.Metadata,
		req.Language,
	)
	if err := s.repo.CreateCollectionBook(collectionID, book); err != nil {
		s.logger.Error(s.ctx, "failed to create collection book", err, map[string]any{
			"user_id":       userID,
			"collection_id": collectionID,
			"name":          req.Name,
		})
		return nil, err
	}
	return book, nil
}

func (s *Service) ensureUniqueSlug(userID, name string) (string, error) {
	base := entity.GenerateSlug(name)
	slug := base
	count := 1
	for {
		_, err := s.repo.FindBookBySlug(userID, slug)
		if err != nil {
			if err == entity.ErrNotFound {
				return slug, nil
			}
			return "", err
		}
		slug = fmt.Sprintf("%s-%d", base, count)
		count++
	}
}
