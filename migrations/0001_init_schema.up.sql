----------------------------------------------------------------------------------------------------
PRAGMA foreign_keys = ON;

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tasks (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text NOT NULL,
	description text,
	date_target DATETIME,
	date_start DATETIME DEFAULT CURRENT_TIMESTAMP,
	archived boolean DEFAULT FALSE
);

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sessions (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	date date,
	duration_mins integer,
	score_feedback integer,
	notes text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS milestones (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	type TEXT,
	value integer,
	achieved date,
	message text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reviews (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	week integer,
	summary text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS config (
	key TEXT PRIMARY KEY,
	value text
);

----------------------------------------------------------------------------------------------------
CREATE TABLE coach (
	id integer PRIMARY KEY AUTOINCREMENT,
	trigger TEXT NOT NULL, -- e.g. 'milestone', 'streak'
	content text NOT NULL,
	date date NULL -- when it was last sent
);

----------------------------------------------------------------------------------------------------
CREATE TABLE calendar (
	id integer PRIMARY KEY AUTOINCREMENT,
	date date NOT NULL,
	note text NOT NULL
);

----------------------------------------------------------------------------------------------------
CREATE TABLE tags (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text UNIQUE NOT NULL
);

----------------------------------------------------------------------------------------------------
CREATE TABLE task_tags (
	task integer NOT NULL REFERENCES tasks (id),
	tag integer NOT NULL REFERENCES tags (id),
	PRIMARY KEY (task, tag)
);

----------------------------------------------------------------------------------------------------
