package collection

import (
	"akira/internal/entity"
	"context"
)

func Make(ctx context.Context, event entity.EventService, logger entity.Logger) entity.CollectionService {
	// todo: add repo
	return NewService(ctx, nil, event, logger)
}
