CREATE TABLE goose_db_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp TIMESTAMP DEFAULT (datetime('now'))
	);
CREATE TABLE sqlite_sequence(name,seq);
CREATE TABLE sessions (
  id TEXT PRIMARY KEY NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
, locale TEXT NOT NULL DEFAULT 'fr');
CREATE TABLE stops (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  code TEXT NOT NULL,
  geo TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  code TEXT NOT NULL,
  destination TEXT NOT NULL,
  direction INT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
, mode TEXT, color TEXT);
CREATE TABLE stops_by_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  stop_id INTEGER NOT NULL,
  line_id INTEGER NOT NULL,
  "order" INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_stop FOREIGN KEY (stop_id) REFERENCES stops(id),
  CONSTRAINT fk_line FOREIGN KEY (line_id) REFERENCES lines(id)
);
CREATE TABLE dashboards (
  id integer primary key autoincrement not null,
  session_id text not null,
  stop_id integer not null,
  created_at datetime default current_timestamp,
  constraint fk_session foreign key (session_id) references sessions(id),
  constraint fk_stop foreign key (stop_id) references stops(id)
);
