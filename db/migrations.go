/*
Copyright © 2025 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package db

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"database/sql"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// migrations holds all DDL statements to initialize or update the schema.
var migrations = []string{
	`CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        description TEXT,
        goal_hours REAL,
        goal_days INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        archived BOOLEAN DEFAULT FALSE
    );`,
	`CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        date DATE NOT NULL,
        duration_minutes INTEGER,
        score INTEGER,
        notes TEXT,
        FOREIGN KEY (task_id) REFERENCES tasks(id)
    );`,
	`CREATE TABLE IF NOT EXISTS milestones (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        type TEXT,
        value INTEGER,
        achieved_at DATE,
        message TEXT,
        FOREIGN KEY (task_id) REFERENCES tasks(id)
    );`,
	`CREATE TABLE IF NOT EXISTS reviews (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        week INTEGER,
        summary TEXT,
        mood TEXT,
        FOREIGN KEY (task_id) REFERENCES tasks(id)
    );`,
	`CREATE TABLE IF NOT EXISTS config (
        key TEXT PRIMARY KEY,
        value TEXT
    );`,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// RunMigrations executes each DDL statement in order. If any Exec fails,
// it returns the error immediately.
func RunMigrations(db *sql.DB) error {
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
