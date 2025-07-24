-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN theme TEXT NOT NULL DEFAULT 'dark';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN theme;
-- +goose StatementEnd
