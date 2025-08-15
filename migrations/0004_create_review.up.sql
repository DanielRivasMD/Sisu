CREATE TABLE IF NOT EXISTS reviews (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	week integer,
	summary text,
	mood text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

