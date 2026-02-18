-- +goose NO TRANSACTION
-- +goose Up
-- Enable Write-Ahead Logging for better concurrency
-- This is stored in the database file, so only needs to be set once
PRAGMA journal_mode = WAL;

-- +goose Down
-- Revert to default delete mode
PRAGMA journal_mode = DELETE;
