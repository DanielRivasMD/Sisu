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
	"reflect"
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

	RegisterCrudSubcommands(calendarCmd, "sisu.db", CrudModel[*models.Calendar]{
		Singular: "calendar",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Calendar, error) {
			return models.Calendars(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(c *models.Calendar) (int64, string) {
			date := c.Date.Format("2006-01-02")
			return c.ID.Int64,
				fmt.Sprintf("date=%s note=%s", date, c.Note)
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			cal, err := models.FindCalendar(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = cal.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runCalendarAdd launches a TUI to create a new calendar entry
func runCalendarAdd(_ *cobra.Command, _ []string) {
	entry := &models.Calendar{}

	fields := []Field{
		{
			Label:   "Date (YYYY-MM-DD)",
			Initial: time.Now().Format("2006-01-02"),
			Parse: func(s string) (interface{}, error) {
				return time.Parse("2006-01-02", s)
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Date").
					Set(reflect.ValueOf(v.(time.Time)))
			},
		},
		{
			Label:   "Note",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return nil, fmt.Errorf("note cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Note").
					SetString(v.(string))
			},
		},
	}

	RunFormWizard(fields, entry)

	if err := entry.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert calendar entry: %v", err)
	}
	fmt.Printf("✅ Created calendar %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runCalendarEdit launches a TUI seeded from an existing record
func runCalendarEdit(_ *cobra.Command, args []string) {
	rawID := args[0]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid calendar ID %q: %v", rawID, err)
	}
	id := null.Int64From(idNum)

	entry, err := models.FindCalendar(context.Background(), db.Conn, id)
	if err != nil {
		log.Fatalf("find calendar %d: %v", id, err)
	}

	fields := []Field{
		{
			Label:   "Date (YYYY-MM-DD)",
			Initial: entry.Date.Format("2006-01-02"),
			Parse: func(s string) (interface{}, error) {
				return time.Parse("2006-01-02", s)
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Date").
					Set(reflect.ValueOf(v.(time.Time)))
			},
		},
		{
			Label:   "Note",
			Initial: entry.Note,
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return nil, fmt.Errorf("note cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Note").
					SetString(v.(string))
			},
		},
	}

	RunFormWizard(fields, entry)

	if _, err := entry.Update(context.Background(), db.Conn, boil.Whitelist("date", "note")); err != nil {
		log.Fatalf("update calendar entry: %v", err)
	}
	fmt.Printf("✅ Updated calendar %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
