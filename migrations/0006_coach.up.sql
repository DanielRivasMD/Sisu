CREATE TABLE coach (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	trigger TEXT NOT NULL, -- e.g. 'milestone', 'streak'
	content TEXT NOT NULL,
	date DATE NULL -- when it was last sent
);

