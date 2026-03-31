-- +goose Up
ALTER TABLE feeds ADD COLUMN last_modified TEXT;
ALTER TABLE feeds ADD COLUMN etag TEXT;

-- +goose Down
ALTER TABLE feeds DROP COLUMN last_modified;
ALTER TABLE feeds DROP COLUMN etag;