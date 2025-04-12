-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS collections (
    id CHAR(26) PRIMARY KEY NOT NULL,
    name VARCHAR(255) NOT NULL,
    edition VARCHAR(255) NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    total_volumes INT NULL,
    user_id CHAR(26) NOT NULL,
    authors TEXT NULL, -- JSON array
    publisher VARCHAR(255) NULL,
    tags TEXT NULL, -- JSON array
    metadata TEXT NULL, -- JSON object
    release_status VARCHAR(255) NULL,
    sync_status VARCHAR(255) NULL,
    sync_sources TEXT NULL, -- JSON array
    crawler_options TEXT NULL, -- JSON object
    lang VARCHAR(255) NULL,
    last_sync_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_collection_user_id ON collections(user_id);
CREATE INDEX IF NOT EXISTS idx_collection_slug ON collections(slug);
CREATE INDEX IF NOT EXISTS idx_collection_sync_status ON collections(sync_status);
CREATE INDEX IF NOT EXISTS idx_collection_last_sync_at ON collections(last_sync_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS collections;
-- +goose StatementEnd
