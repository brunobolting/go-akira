package crawler

import (
	"akira/internal/entity"
	"context"
	"time"
)

var _ entity.CrawlerConsumer = (*Consumer)(nil)

type Consumer struct {
	service    entity.CrawlerService
	collection entity.CollectionService
	book       entity.BookService
	sub        *entity.Subscriber
	logger     entity.Logger
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewConsumer(
	ctx context.Context,
	service entity.CrawlerService,
	collection entity.CollectionService,
	book entity.BookService,
	event entity.EventService,
	logger entity.Logger,
) *Consumer {
	consumerCtx, cancel := context.WithCancel(ctx)
	consumer := &Consumer{
		service:    service,
		collection: collection,
		book:       book,
		logger:     logger,
		ctx:        consumerCtx,
		cancelFunc: cancel,
	}
	consumer.sub = event.Subscribe(
		"crawler-consumer",
		cancel,
		entity.EventCollectionCreated,
		entity.EventCollectionSyncFetching,
		entity.EventCrawlerCompleted,
		entity.EventCrawlerItemFounded,
	)
	go consumer.ConsumeEvents()
	return consumer
}

func (c *Consumer) ConsumeEvents() {
	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info(c.ctx, "crawler consumer shutting down", nil)
			return
		case event, ok := <-c.sub.Ch:
			if !ok {
				c.logger.Info(c.ctx, "crawler consumer channel closed", nil)
				return
			}
			go func(event entity.Event) {
				defer func() {
					if r := recover(); r != nil {
						c.logger.Error(c.ctx, "panic in crawler consumer", nil, map[string]any{
							"event_type": event.Type,
							"recover":    r,
						})
					}
				}()

				switch event.Type {
				case entity.EventCollectionCreated:
					c.logger.Info(c.ctx, "collection created event received", map[string]any{
						"collection_id": event.Data,
					})
					c.handleCollectionCreated(event)
				case entity.EventCollectionSyncFetching:
					// Handle the sync fetching event
				case entity.EventCrawlerCompleted:
					// Handle the crawler completed event
				case entity.EventCrawlerItemFounded:
					c.logger.Info(c.ctx, "crawler item founded event received", map[string]any{
						"collection_id": event.Data,
					})
					c.handleCrawlerItemFounded(event)
				}
			}(event)
		}
	}
}

func (c *Consumer) Shutdown() error {
	c.cancelFunc()
	c.sub.Cancel()
	return nil
}

func (c *Consumer) handleCollectionCreated(event entity.Event) {
	data, ok := event.Data.(*entity.Collection)
	if !ok {
		c.logger.Warn(c.ctx, "invalid event data typpe", map[string]any{
			"expected": "*entity.Collection",
			"received": event.Data,
		})
		return
	}
	if !data.CrawlerOptions.AutoSync || len(data.SyncSources) == 0 {
		c.logger.Info(c.ctx, "collection not configured for auto sync", map[string]any{
			"collection_id": data.ID,
		})
		return
	}
	if !data.CrawlerOptions.TrackNewVolumes {
		c.logger.Info(c.ctx, "collection not configured to track new volumes", map[string]any{
			"collection_id": data.ID,
		})
		return
	}
	searchTerms := []string{data.Name}
	if data.Metadata != nil {
		if terms, ok := data.Metadata["search_terms"]; ok {
			searchTerms = append(searchTerms, terms)
		}
	}
	// searchTerms = append(searchTerms, data.Edition)
	searchTerms = append(searchTerms, data.Author...)

	opts := entity.CrawlerOptions{
		MaxPages:        50,
		Timeout:         3 * time.Minute,
		MaxConcurrency:  2,
		RequestInterval: 3 * time.Second,
	}

	req := entity.CrawlerRequest{
		CollectionID: data.ID,
		SearchTerms:  searchTerms,
		Sites:        data.SyncSources,
		Opts:         opts,
	}
	c.logger.Info(c.ctx, "starting crawler", map[string]any{
		"collection_id": data.ID,
		"search_terms":  searchTerms,
		"sites":         data.SyncSources,
	})
	if err := c.service.FetchCollection(c.ctx, req); err != nil {
		c.logger.Error(c.ctx, "failed to start crawler", err, map[string]any{
			"collection_id": data.ID,
		})
	}
}

func (c *Consumer) handleCrawlerItemFounded(event entity.Event) {
	data, ok := event.Data.(map[string]any)
	if !ok {
		c.logger.Warn(c.ctx, "invalid event data type", map[string]any{
			"expected": "map[string]any",
			"received": event.Data,
		})
		return
	}
	collectionID, ok := data["collection_id"].(string)
	if !ok || collectionID == "" {
		c.logger.Warn(c.ctx, "invalid collection ID", nil)
		return
	}
	result, ok := data["result"].(entity.CrawledResult)
	if !ok {
		c.logger.Warn(c.ctx, "invalid crawled result", nil)
		return
	}
	c.logger.Debug(c.ctx, "crawled item founded", map[string]any{
		"collection_id": collectionID,
		"title":         result.Title,
		"volume":        result.Volume,
		"isbn":          result.ISBN,
		"price":         result.Price,
		"source":        result.Source,
	})
}
