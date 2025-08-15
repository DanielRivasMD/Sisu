CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task INTEGER NOT NULL,
	date DATE NOT NULL,
	duration_mins INTEGER,
	score_feedback INTEGER,
	notes TEXT,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

