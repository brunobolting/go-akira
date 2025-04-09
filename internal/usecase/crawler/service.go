package crawler

import (
	"akira/internal/entity"
	"akira/internal/usecase/crawler/provider"
	"context"
	"sync"
	"time"
)

var _ entity.CrawlerService = (*Service)(nil)

type Service struct {
	providers      map[string]entity.SiteProvider
	event          entity.EventService
	logger         entity.Logger
	ctx            context.Context
	activeCrawlers sync.Map
	results        sync.Map
}

func NewService(
	ctx context.Context,
	event entity.EventService,
	logger entity.Logger,
) *Service {
	service := &Service{
		providers: make(map[string]entity.SiteProvider),
		event:     event,
		logger:    logger,
		ctx:       ctx,
	}
	service.RegisterProvider(provider.NewAmazonProvider())
	service.RegisterProvider(provider.NewPaniniProvider())
	return service
}

func (s *Service) RegisterProvider(provider entity.SiteProvider) {
	s.providers[provider.SiteName()] = provider
}

func (s *Service) FetchCollection(ctx context.Context, req entity.CrawlerRequest) error {
	if _, exists := s.activeCrawlers.Load(req.CollectionID); exists {
		return entity.ErrCrawlerAlreadyRunning
	}
	s.activeCrawlers.Store(req.CollectionID, entity.SyncStatusFetching)
	go func() {
		crawlerCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		s.activeCrawlers.Store(req.CollectionID, cancel)
		s.event.Publish(entity.NewEvent(
			entity.EventCrawlerStarted,
			"",
			map[string]any{
				"collection_id": req.CollectionID,
				"search_terms":  req.SearchTerms,
				"sites":         req.Sites,
			},
		))
		s.logger.Info(crawlerCtx, "crawler started", map[string]any{
			"collection_id": req.CollectionID,
			"search_terms":  req.SearchTerms,
			"sites":         req.Sites,
		})
		resultsChan := make(chan entity.CrawledResult, 100)
		errChan := make(chan error, len(req.Sites))
		var wg sync.WaitGroup
		for _, site := range req.Sites {
			provider, ok := s.providers[site]
			if !ok {
				s.logger.Warn(crawlerCtx, "provider not found", map[string]any{
					"site": site,
				})
				continue
			}
			wg.Add(1)
			go func(site string, provider entity.SiteProvider) {
				defer wg.Done()
				err := provider.Setup(req.Opts)
				if err != nil {
					s.logger.Error(crawlerCtx, "provider setup failed", err, map[string]any{
						"site": site,
					})
					errChan <- err
					return
				}
				results, err := provider.Fetch(crawlerCtx, req.SearchTerms)
				if err != nil {
					s.logger.Error(crawlerCtx, "provider fetch failed", err, map[string]any{
						"site": site,
					})
					errChan <- err
					return
				}
				for _, result := range results {
					select {
					case resultsChan <- result:
						s.event.Publish(entity.NewEvent(
							entity.EventCrawlerItemFounded,
							"",
							map[string]any{
								"collection_id": req.CollectionID,
								"site":          site,
								"result":        result,
							},
						))
					case <-crawlerCtx.Done():
						s.logger.Info(crawlerCtx, "crawler context done", map[string]any{
							"site": site,
						})
						return
					}
				}
			}(site, provider)
		}
		go func() {
			wg.Wait()
			close(resultsChan)
			close(errChan)
			s.logger.Debug(crawlerCtx, "all providers finished", nil)
		}()
		var results []entity.CrawledResult
		for result := range resultsChan {
			results = append(results, result)
			s.logger.Debug(crawlerCtx, "received result", map[string]any{
				"title":  result.Title,
				"volume": result.Volume,
				"source": result.Source,
			})
		}

		s.logger.Info(crawlerCtx, "collected results from channel", map[string]any{
			"result_count": len(results),
		})

		var errors []error
		for err := range errChan {
			errors = append(errors, err)
			s.logger.Debug(crawlerCtx, "received error", map[string]any{
				"error": err.Error(),
			})
		}
		s.results.Store(req.CollectionID, results)
		if len(errors) > 0 && len(results) == 0 {
			s.activeCrawlers.Store(req.CollectionID, entity.SyncStatusFailed)
			s.event.Publish(entity.NewEvent(
				entity.EventCrawlerFailed,
				"",
				map[string]any{
					"collection_id": req.CollectionID,
					"errors":        errors,
				},
			))
			s.logger.Error(crawlerCtx, "crawler failed", nil, map[string]any{
				"collection_id": req.CollectionID,
				"errors":        errors,
			})
			return
		}
		s.activeCrawlers.Store(req.CollectionID, entity.SyncStatusSynced)
		s.event.Publish(entity.NewEvent(
			entity.EventCrawlerCompleted,
			"",
			map[string]any{
				"collection_id": req.CollectionID,
				"result_count":  len(results),
				"has_errors":    len(errors) > 0,
			},
		))
		s.logger.Info(crawlerCtx, "crawler completed", map[string]any{
			"collection_id": req.CollectionID,
			"result_count":  len(results),
			"error_count":   len(errors),
		})
		s.logger.Debug(crawlerCtx, "crawler results", map[string]any{
			"collection_id": req.CollectionID,
			"results":       results,
			"errors":        errors,
		})
	}()
	return nil
}

func (s *Service) GetStatus(collectionID string) (entity.SyncStatus, error) {
	status, exists := s.activeCrawlers.Load(collectionID)
	if !exists {
		return entity.SyncStatusNotFound, nil
	}
	if status, ok := status.(entity.SyncStatus); ok {
		return status, nil
	}
	return entity.SyncStatusFetching, nil
}

func (s *Service) CancelFetch(collectionID string) error {
	value, exists := s.activeCrawlers.Load(collectionID)
	if !exists {
		return entity.ErrCrawlerNotRunning
	}
	if cancelFunc, ok := value.(context.CancelFunc); ok {
		cancelFunc()
		s.activeCrawlers.Store(collectionID, entity.SyncStatusFailed)
		s.event.Publish(entity.NewEvent(
			entity.EventCrawlerFailed,
			"",
			map[string]any{
				"collection_id": collectionID,
				"reason":        "canceled by user",
			},
		))
		return nil
	}
	return entity.ErrCrawlerCannotBeCancelled
}
