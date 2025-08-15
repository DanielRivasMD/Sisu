CREATE TABLE IF NOT EXISTS milestones (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	type TEXT,
	value integer,
	achieved date,
	message text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

