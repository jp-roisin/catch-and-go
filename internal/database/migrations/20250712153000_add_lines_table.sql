-- +goose Up
-- +goose StatementBegin
CREATE TABLE lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  code TEXT NOT NULL,
  destination TEXT NOT NULL,
  direction INT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lines;
-- +goose StatementEnd
