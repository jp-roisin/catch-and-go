-- +goose Up
-- +goose StatementBegin
CREATE TABLE stops (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  code TEXT NOT NULL,
  geo TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stops;
-- +goose StatementEnd
