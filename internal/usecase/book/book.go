package book

import (
	"akira/internal/entity"
	"context"
)

func Make(ctx context.Context, logger entity.Logger) entity.BookService {
	repo := NewMemoRepository()
	return NewService(ctx, repo, logger)
}
