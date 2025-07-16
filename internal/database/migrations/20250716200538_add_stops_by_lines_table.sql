-- +goose Up
-- +goose StatementBegin
CREATE TABLE stops_by_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  stop_id INTEGER NOT NULL,
  line_id INTEGER NOT NULL,
  "order" INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_stop FOREIGN KEY (stop_id) REFERENCES stops(id),
  CONSTRAINT fk_line FOREIGN KEY (line_id) REFERENCES lines(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stops_by_lines;
-- +goose StatementEnd
