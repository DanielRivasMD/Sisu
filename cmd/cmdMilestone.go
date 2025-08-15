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

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var milestoneCmd = &cobra.Command{
	Use:     "milestone",
	Short:   "",
	Long:    helpMilestone,
	Example: exampleMilestone,

	// Run: runMilestone,

}

////////////////////////////////////////////////////////////////////////////////////////////////////

var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	// rootCmd.AddCommand(milestoneCmd)

	rootCmd.AddCommand(
		NewCrudCmd("sisu.db", CrudModel[*models.Milestone]{
			Singular: "milestone",

			// 1. ListFn returns all milestones
			ListFn: func(ctx context.Context, db *sql.DB) ([]*models.Milestone, error) {
				return models.Milestones(
					qm.OrderBy("id ASC"),
				).All(ctx, db)
			},

			// 2. Format for display in "list"
			Format: func(m *models.Milestone) (int64, string) {
				// Task is an int64
				taskID := m.Task

				// Type, Value, Message, Achieved are nullable wrappers
				typ := m.Type.String
				value := m.Value.Int64

				// Only show date if not zero
				ach := ""
				if !m.Achieved.Time.IsZero() {
					ach = m.Achieved.Time.Format("2006-01-02")
				}

				msg := m.Message.String

				return m.ID.Int64, fmt.Sprintf(
					"task=%d type=%s value=%d achieved=%s msg=%s",
					taskID, typ, value, ach, msg,
				)
			},

			// 4. RemoveFn deletes by PK
			RemoveFn: func(ctx context.Context, db *sql.DB, id int64) error {
				m, err := models.FindMilestone(ctx, db, null.Int64From(id))
				if err != nil {
					return err
				}
				_, err = m.Delete(ctx, db)
				return err
			},
		}),
	)

}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpMilestone = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

var exampleMilestone = formatExample(
	"",
	[]string{"milestone"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// func runMilestone(cmd *cobra.Command, args []string) {

// }

////////////////////////////////////////////////////////////////////////////////////////////////////
