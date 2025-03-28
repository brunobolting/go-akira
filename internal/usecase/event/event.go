package event

import (
	"akira/internal/entity"
	"context"
)

func Make(ctx context.Context, logger entity.Logger) entity.EventService {
	return NewService(ctx, logger)
}
