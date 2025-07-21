-- +goose Up
-- +goose StatementBegin
create table dashboards (
  id integer primary key autoincrement not null,
  session_id text not null,
  stop_id integer not null,
  created_at datetime default current_timestamp,
  constraint fk_session foreign key (session_id) references sessions(id),
  constraint fk_stop foreign key (stop_id) references stops(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dashboards;
-- +goose StatementEnd
