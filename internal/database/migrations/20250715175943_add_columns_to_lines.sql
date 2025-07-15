-- +goose Up
-- +goose StatementBegin
ALTER TABLE lines ADD COLUMN mode TEXT;
ALTER TABLE lines ADD COLUMN color TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE lines DROP COLUMN mode;
ALTER TABLE lines DROP COLUMN color;
-- +goose StatementEnd
