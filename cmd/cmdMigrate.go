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
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "Initialize or migrate the database schema",
	Long:    helpMigrate,
	Example: exampleMigrate,

	Run: runMigrate,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(migrateCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpMigrate = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Runs every pending .up.sql migration against your SQLite database"+
		"If the database file doesn’t exist yet, it will be created automatically",
)

var exampleMigrate = formatExample(
	"",
	[]string{"migrate"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func runMigrate(cmd *cobra.Command, args []string) {

	// InitDB creates the file and runs migrations
	conn, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("❌ migrate failed: %v", err)
	}

	fmt.Printf("✅ migrations applied; database at %s\n", dbPath)

	if conn != nil {
		if err := conn.Close(); err != nil {
			log.Printf("⚠️ error closing DB: %v", err)
		}
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////
