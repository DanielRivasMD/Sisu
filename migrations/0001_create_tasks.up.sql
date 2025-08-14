CREATE TABLE IF NOT EXISTS tasks (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    description   TEXT,
    goal_hours    REAL,
    goal_days     INTEGER,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    archived      BOOLEAN DEFAULT FALSE
);
