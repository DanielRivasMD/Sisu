CREATE TABLE IF NOT EXISTS reviews (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    week    INTEGER,
    summary TEXT,
    mood    TEXT,
    FOREIGN KEY(task_id) REFERENCES tasks(id)
);
