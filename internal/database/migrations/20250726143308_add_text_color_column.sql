-- +goose Up
-- +goose StatementBegin
ALTER TABLE lines ADD COLUMN text_color TEXT NOT NULL DEFAULT '#ffffff';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE lines DROP COLUMN text_color;
-- +goose StatementEnd
