CREATE TABLE IF NOT EXISTS tasks (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text NOT NULL,
	description text,
	days integer,
	created DATETIME DEFAULT CURRENT_TIMESTAMP,
	archived boolean DEFAULT FALSE
);

