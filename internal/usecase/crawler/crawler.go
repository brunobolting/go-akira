package crawler

import (
	"akira/internal/entity"
	"context"
)

func Make(
	ctx context.Context,
	event entity.EventService,
	book entity.BookService,
	collection entity.CollectionService,
	logger entity.Logger,
) (entity.CrawlerService, entity.CrawlerConsumer) {
	service := NewService(ctx, event, logger)
	consumer := NewConsumer(ctx, service, collection, book, event, logger)
	return service, consumer
}
