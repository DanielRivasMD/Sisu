CREATE TABLE task_tags (
	task integer NOT NULL REFERENCES tasks (id),
	tag integer NOT NULL REFERENCES tags (id),
	PRIMARY KEY (task, tag)
);

