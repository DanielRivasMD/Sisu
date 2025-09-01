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

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// Schema (coach):
//   id INTEGER PK
//   trigger TEXT NOT NULL   → string (required)
//   content TEXT NOT NULL   → string (required)
//   date DATE               → null.Time (optional)

var coachCmd = &cobra.Command{
	Use:   "coach",
	Short: "Manage coach triggers (list, rm via CLI; add/edit via TUI)",
}

var coachAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new coach entry",
	Run:   runCoachAdd,
}

var coachEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit a coach entry",
	Args:  cobra.ExactArgs(1),
	Run:   runCoachEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(coachCmd)
	coachCmd.AddCommand(coachAddCmd, coachEditCmd)

	RegisterCrudSubcommands(coachCmd, "sisu.db", CrudModel[*models.Coach]{
		Singular: "coach",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Coach, error) {
			return models.Coaches(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(c *models.Coach) (int64, string) {
			date := ""
			if c.Date.Valid {
				date = c.Date.Time.Format("2006-01-02")
			}
			return c.ID.Int64, fmt.Sprintf("trigger=%s content=%s date=%s", c.Trigger, c.Content, date)
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			c, err := models.FindCoach(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = c.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runCoachAdd(_ *cobra.Command, _ []string) {
	entry := &models.Coach{}

	fields := []Field{
		FString("Trigger", "Trigger", ""),
		FString("Content", "Content", ""),
		// date is optional in the schema → null.Time
		FOptDate("Date (YYYY-MM-DD, optional)", "Date", ""),
	}

	RunFormWizard(fields, entry)

	if err := entry.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert coach entry: %v", err)
	}
	fmt.Printf("Created coach %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runCoachEdit(_ *cobra.Command, args []string) {
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid coach ID: %v", err)
	}

	entry, err := models.FindCoach(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find coach %d: %v", idNum, err)
	}

	fields := []Field{
		FString("Trigger", "Trigger", entry.Trigger),
		FString("Content", "Content", entry.Content),
		FOptDate("Date (YYYY-MM-DD, optional)", "Date",
			func() string {
				if entry.Date.Valid {
					return entry.Date.Time.Format("2006-01-02")
				}
				return ""
			}(),
		),
	}

	RunFormWizard(fields, entry)

	if _, err := entry.Update(context.Background(), db.Conn, boil.Whitelist("trigger", "content", "date")); err != nil {
		log.Fatalf("update coach entry: %v", err)
	}
	fmt.Printf("Updated coach %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
