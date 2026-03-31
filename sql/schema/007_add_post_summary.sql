-- +goose Up
ALTER TABLE posts ADD COLUMN summary TEXT;

-- +goose Down
ALTER TABLE posts DROP COLUMN summary;