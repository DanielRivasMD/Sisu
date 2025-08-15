CREATE TABLE IF NOT EXISTS sessions (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	date date NOT NULL,
	duration integer,
	score integer,
	notes text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

