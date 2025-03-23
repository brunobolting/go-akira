-- +goose Up
-- +goose no transaction
PRAGMA auto_vacuum = incremental;
PRAGMA journal_mode = WAL;
PRAGMA page_size = 32768;
