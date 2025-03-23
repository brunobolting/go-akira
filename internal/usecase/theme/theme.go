package theme

import (
	"akira/internal/entity"
	"context"
)

func Make(ctx context.Context, logger entity.Logger) *Service {
	return NewService(ctx, logger)
}
