-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
  id TEXT PRIMARY KEY NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
