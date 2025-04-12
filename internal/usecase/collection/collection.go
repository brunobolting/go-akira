package collection

import (
	"akira/internal/entity"
	"context"
	"database/sql"
)

func Make(ctx context.Context, db *sql.DB, event entity.EventService, logger entity.Logger) entity.CollectionService {
	repo := NewCollectionSqliteRepository(db)
	return NewService(ctx, repo, event, logger)
}
