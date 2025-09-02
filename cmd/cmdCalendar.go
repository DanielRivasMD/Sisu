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
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// Schema (calendar):
//   id INTEGER PK
//   date DATE            → null.Time (nullable)
//   note TEXT NOT NULL   → string

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Manage calendar notes (list, rm via CLI; add/edit via TUI)",
}

var calendarAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new calendar note",
	Run:   runCalendarAdd,
}

var calendarEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit a calendar note",
	Args:  cobra.ExactArgs(1),
	Run:   runCalendarEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(calendarCmd)
	calendarCmd.AddCommand(calendarAddCmd, calendarEditCmd)

	// cmd/cmdCalendar.go (inside init(): RegisterCrudSubcommands for calendar)

	RegisterCrudSubcommands(calendarCmd, "sisu.db", CrudModel[*models.Calendar]{
		Singular: "calendar",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Calendar, error) {
			return models.Calendars(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(c *models.Calendar) (int64, string) {
			date := ""
			if c.Date.Valid {
				date = c.Date.Time.Format(time.RFC3339)
			}
			return c.ID.Int64, fmt.Sprintf("date=%s note=%s", date, c.Note)
		},

		TableHeaders: []string{"id", "date", "note"},
		TableRow: func(c *models.Calendar) []string {
			date := ""
			if c.Date.Valid {
				date = c.Date.Time.Format(time.RFC3339)
			}
			return []string{
				strconv.FormatInt(c.ID.Int64, 10),
				date,
				c.Note,
			}
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			row, err := models.FindCalendar(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = row.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runCalendarAdd(_ *cobra.Command, _ []string) {
	entry := &models.Calendar{}

	fields := []Field{
		FOptDate("Date (YYYY-MM-DD, optional)", "Date", ""),
		FString("Note", "Note", ""),
	}

	RunFormWizard(fields, entry)

	if err := entry.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert calendar entry: %v", err)
	}
	fmt.Printf("Created calendar %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runCalendarEdit(_ *cobra.Command, args []string) {
	rawID := args[0]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid calendar ID %q: %v", rawID, err)
	}

	entry, err := models.FindCalendar(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find calendar %d: %v", idNum, err)
	}

	fields := []Field{
		FOptDate("Date (YYYY-MM-DD, optional)", "Date", OptTimeInitial(entry.Date, "2006-01-02")),
		FString("Note", "Note", entry.Note),
	}

	RunFormWizard(fields, entry)

	if _, err := entry.Update(context.Background(), db.Conn, boil.Whitelist("date", "note")); err != nil {
		log.Fatalf("update calendar entry: %v", err)
	}
	fmt.Printf("Updated calendar %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
