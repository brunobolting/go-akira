package collection

import (
	"akira/internal/entity"
	"context"
	"fmt"
)

var _ entity.CollectionService = (*Service)(nil)

type Service struct {
	repo   entity.CollectionRepository
	event  entity.EventService
	ctx    context.Context
	logger entity.Logger
}

func NewService(ctx context.Context, repo entity.CollectionRepository, event entity.EventService, logger entity.Logger) *Service {
	return &Service{
		repo:   repo,
		event:  event,
		ctx:    ctx,
		logger: logger,
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
	s.event.Publish(entity.NewEvent(
		entity.EventCollectionCreated,
		userID,
		collection,
	))
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
