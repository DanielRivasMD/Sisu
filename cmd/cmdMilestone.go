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

// Schema (milestones):
//   id INTEGER PK
//   task INTEGER NOT NULL   → int64 (required)
//   type TEXT               → null.String (optional)
//   value INTEGER           → null.Int64 (optional)
//   done DATE               → null.Time (optional)
//   message TEXT            → null.String (optional)

var milestoneCmd = &cobra.Command{
	Use:               "milestone",
	Short:             "Manage milestones",
	Long:              helpMilestone,
	Example:           exampleMilestone,
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var milestoneAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new milestone",
	Run:   runMilestoneAdd,
}

var milestoneEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit a milestone",
	Args:  cobra.ExactArgs(1),
	Run:   runMilestoneEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(milestoneCmd)
	milestoneCmd.AddCommand(milestoneAddCmd, milestoneEditCmd)

	RegisterCrudSubcommands(milestoneCmd, "sisu.db", CrudModel[*models.Milestone]{
		Singular: "milestone",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Milestone, error) {
			return models.Milestones(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(m *models.Milestone) (int64, string) {
			done := ""
			if m.Done.Valid {
				done = m.Done.Time.Format(time.RFC3339)
			}
			val := ""
			if m.Value.Valid {
				val = strconv.FormatInt(m.Value.Int64, 10)
			}
			return m.ID.Int64, fmt.Sprintf("task=%d type=%s value=%s done=%s msg=%s",
				m.Task, m.Type.String, val, done, m.Message.String)
		},

		TableHeaders: []string{"id", "task", "type", "value", "done", "message"},
		TableRow: func(m *models.Milestone) []string {
			done := ""
			if m.Done.Valid {
				done = m.Done.Time.Format(time.RFC3339)
			}
			val := ""
			if m.Value.Valid {
				val = strconv.FormatInt(m.Value.Int64, 10)
			}
			return []string{
				strconv.FormatInt(m.ID.Int64, 10),
				strconv.FormatInt(m.Task, 10),
				m.Type.String,
				val,
				done,
				m.Message.String,
			}
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			row, err := models.FindMilestone(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = row.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runMilestoneAdd(_ *cobra.Command, _ []string) {
	m := &models.Milestone{}

	fields := []Field{
		FInt("Task ID", "Task", ""),
		FOptString("Type (optional)", "Type", ""),
		FOptInt("Value (optional)", "Value", ""),
		FOptDate("Done date (YYYY-MM-DD, optional)", "Done", ""),
		FOptString("Message (optional)", "Message", ""),
	}

	RunFormWizard(fields, m)

	if err := m.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert milestone: %v", err)
	}
	fmt.Printf("Created milestone %d\n", m.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runMilestoneEdit(_ *cobra.Command, args []string) {
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid milestone ID %q: %v", args[0], err)
	}

	m, err := models.FindMilestone(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find milestone %d: %v", idNum, err)
	}

	fields := []Field{
		FInt("Task ID", "Task", strconv.FormatInt(m.Task, 10)),
		FOptString("Type (optional)", "Type", m.Type.String),
		FOptInt("Value (optional)", "Value", OptInt64Initial(m.Value)),
		FOptDate("Done date (YYYY-MM-DD, optional)", "Done", OptTimeInitial(m.Done, DateYMD)),
		FOptString("Message (optional)", "Message", m.Message.String),
	}

	RunFormWizard(fields, m)

	if _, err := m.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update milestone: %v", err)
	}
	fmt.Printf("Updated milestone %d\n", m.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
