-- +goose Up
-- Add a GIN index to the title column for fast Full-Text Search
CREATE INDEX posts_title_fts_idx ON posts USING GIN (to_tsvector('english', title));

-- +goose Down
DROP INDEX posts_title_fts_idx;