CREATE TABLE IF NOT EXISTS milestones (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id     INTEGER NOT NULL,
    type        TEXT,
    value       INTEGER,
    achieved_at DATE,
    message     TEXT,
    FOREIGN KEY(task_id) REFERENCES tasks(id)
);
