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

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// InitDB opens (and if needed, creates) the SQLite database at the given path,
// runs migrations to ensure the schema is up to date, and returns the *sql.DB.
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Run all migrations (CREATE TABLE IF NOT EXISTS …)
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
