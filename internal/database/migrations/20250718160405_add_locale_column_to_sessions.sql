-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN locale TEXT NOT NULL DEFAULT 'fr';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN locale;
-- +goose StatementEnd
