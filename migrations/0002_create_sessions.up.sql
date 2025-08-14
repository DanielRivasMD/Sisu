CREATE TABLE IF NOT EXISTS sessions (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id          INTEGER NOT NULL,
    date             DATE    NOT NULL,
    duration_minutes INTEGER,
    score            INTEGER,
    notes            TEXT,
    FOREIGN KEY(task_id) REFERENCES tasks(id)
);
