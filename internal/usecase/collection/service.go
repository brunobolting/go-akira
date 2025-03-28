package collection

import (
	"akira/internal/entity"
	"fmt"
)

var _ entity.CollectionService = (*Service)(nil)

type Service struct {
	repo entity.CollectionRepository
}

func NewService(repo entity.CollectionRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateCollection(userID string, req entity.CreateCollectionRequest) (*entity.Collection, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	slug, err := s.ensureUniqueSlug(userID, req.Name)
	if err != nil {
		return nil, err
	}
	collection := entity.NewCollection(
		userID,
		req.Name,
		req.Edition,
		slug,
		req.Author,
		req.Publisher,
		req.Language,
		req.Tags,
		req.Metadata,
		req.CrawlerDataSource,
		req.CrawlerOptions,
	)
	if err := s.repo.CreateCollection(collection); err != nil {
		return nil, err
	}
	// todo: send event and queue sync
	return collection, nil
}

func (s *Service) ensureUniqueSlug(userID, name string) (string, error) {
	base := entity.GenerateSlug(name)
	slug := base
	count := 1
	for {
		_, err := s.repo.FindCollectionBySlug(userID, slug)
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
