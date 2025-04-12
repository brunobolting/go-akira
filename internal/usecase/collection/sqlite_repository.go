package collection

import (
	"akira/internal/entity"
	"database/sql"
	"encoding/json"
)

var _ entity.CollectionRepository = (*CollectionSqliteRepository)(nil)

type CollectionSqliteRepository struct {
	db *sql.DB
}

func NewCollectionSqliteRepository(db *sql.DB) entity.CollectionRepository {
	return &CollectionSqliteRepository{db: db}
}

func (r *CollectionSqliteRepository) scanCollectionRow(row *sql.Row) (*entity.Collection, error) {
	var collection entity.Collection
	var nullableEdition, nullableAuthor, nullableTags, nullableMetadata, nullableSyncSources, nullableCrawlerOptions sql.NullString
	var nullablePublisher, nullableLang sql.NullString
	var nullableTotalVolumes sql.NullInt32
	var nullableLastSync sql.NullTime
	err := row.Scan(
		&collection.ID,
		&collection.Name,
		&nullableEdition,
		&collection.Slug,
		&collection.UserID,
		&nullableAuthor,
		&nullablePublisher,
		&nullableTags,
		&nullableMetadata,
		&collection.ReleaseStatus,
		&collection.SyncStatus,
		&nullableSyncSources,
		&nullableTotalVolumes,
		&nullableCrawlerOptions,
		&nullableLang,
		&nullableLastSync,
		&collection.CreatedAt,
		&collection.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	if nullableAuthor.Valid {
		err = json.Unmarshal([]byte(nullableAuthor.String), &collection.Author)
		if err != nil {
			return nil, err
		}
	}
	if nullableTags.Valid {
		err = json.Unmarshal([]byte(nullableTags.String), &collection.Tags)
		if err != nil {
			return nil, err
		}
	}
	if nullableMetadata.Valid {
		err = json.Unmarshal([]byte(nullableMetadata.String), &collection.Metadata)
		if err != nil {
			return nil, err
		}
	}
	if nullableSyncSources.Valid {
		err = json.Unmarshal([]byte(nullableSyncSources.String), &collection.SyncSources)
		if err != nil {
			return nil, err
		}
	}
	if nullableCrawlerOptions.Valid {
		err = json.Unmarshal([]byte(nullableCrawlerOptions.String), &collection.CrawlerOptions)
		if err != nil {
			return nil, err
		}
	}
	if nullablePublisher.Valid {
		collection.Publisher = nullablePublisher.String
	}
	if nullableEdition.Valid {
		collection.Edition = nullableEdition.String
	}
	if nullableTotalVolumes.Valid {
		collection.TotalVolumes = int(nullableTotalVolumes.Int32)
	}
	if nullableLang.Valid {
		collection.Language = nullableLang.String
	}
	if nullableLastSync.Valid {
		collection.LastSync = nullableLastSync.Time
	}
	return &collection, nil
}

func (r *CollectionSqliteRepository) CreateCollection(collection *entity.Collection) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO collections (
			id, name, edition, slug, user_id, authors, publisher,
			tags, metadata, release_status, sync_status, sync_sources,
			total_volumes, crawler_options, lang, last_sync_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	authors, err := json.Marshal(collection.Author)
	if err != nil {
		return err
	}
	tags, err := json.Marshal(collection.Tags)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(collection.Metadata)
	if err != nil {
		return err
	}
	syncSources, err := json.Marshal(collection.SyncSources)
	if err != nil {
		return err
	}
	crawlerOptions, err := json.Marshal(collection.CrawlerOptions)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		collection.ID,
		collection.Name,
		collection.Edition,
		collection.Slug,
		collection.UserID,
		authors, // marshal to JSON
		collection.Publisher,
		tags,     // marshal to JSON
		metadata, // marshal to JSON
		collection.ReleaseStatus,
		collection.SyncStatus,
		syncSources, // marshal to JSON
		collection.TotalVolumes,
		crawlerOptions, // marshal to JSON
		collection.Language,
		collection.LastSync,
		collection.CreatedAt,
		collection.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *CollectionSqliteRepository) FindCollectionBySlug(userID, slug string) (*entity.Collection, error) {
	stmt, err := r.db.Prepare(`
		SELECT id, name, edition, slug, user_id, authors, publisher,
			tags, metadata, release_status, sync_status, sync_sources,
			total_volumes, crawler_options, lang, last_sync_at,
			created_at, updated_at
		FROM collections WHERE user_id = ? AND slug = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(userID, slug)
	return r.scanCollectionRow(row)
}
