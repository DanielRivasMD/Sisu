----------------------------------------------------------------------------------------------------
PRAGMA foreign_keys = ON;

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tasks (
	id integer PRIMARY KEY AUTOINCREMENT,
	name text NOT NULL,
	tag text,
	description text,
	start datetime,
	target datetime,
	archived boolean DEFAULT FALSE
);

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sessions (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	date date,
	mins integer,
	feedback integer,
	notes text,
	FOREIGN KEY (task) REFERENCES tasks (id)
);

----------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS milestones (
	id integer PRIMARY KEY AUTOINCREMENT,
	task integer NOT NULL,
	type text,
	value integer,
	done date,
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
CREATE TABLE coach (
	id integer PRIMARY KEY AUTOINCREMENT,
	trigger text NOT NULL,
	content text NOT NULL,
	date date
);

----------------------------------------------------------------------------------------------------
CREATE TABLE calendar (
	id integer PRIMARY KEY AUTOINCREMENT,
	date date,
	note text NOT NULL
);

----------------------------------------------------------------------------------------------------
