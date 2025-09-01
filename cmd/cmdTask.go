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

// Schema (tasks):
//   - id integer PK
//   - name text NOT NULL            → string
//   - tag text                      → null.String
//   - description text              → null.String
//   - target datetime               → null.Time
//   - start datetime                → null.Time
//   - archived boolean DEFAULT FALSE → null.Bool (likely; confirm after regen)

var taskCmd = &cobra.Command{
	Use:               "task",
	Short:             "Manage tasks",
	Long:              helpTask,
	Example:           exampleTask,
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var taskAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new task",
	Run:   runTaskAdd,
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit a task",
	Args:  cobra.ExactArgs(1),
	Run:   runTaskEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskAddCmd, taskEditCmd)

	RegisterCrudSubcommands(taskCmd, "sisu.db", CrudModel[*models.Task]{
		Singular: "task",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Task, error) {
			return models.Tasks(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(t *models.Task) (int64, string) {
			// optional fields
			tag := t.Tag.String
			desc := t.Description.String
			target := ""
			if t.Target.Valid {
				target = t.Target.Time.Format(time.RFC3339)
			}
			start := ""
			if t.Start.Valid {
				start = t.Start.Time.Format(time.RFC3339)
			}
			return t.ID.Int64, fmt.Sprintf(
				"name=%s tag=%s target=%s start=%s archived=%v desc=%s",
				t.Name, tag, target, start, t.Archived.Bool, desc,
			)
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			task, err := models.FindTask(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = task.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskAdd(_ *cobra.Command, _ []string) {
	task := &models.Task{}

	fields := []Field{
		// name (required)
		{
			Label:    "Task name",
			Initial:  "",
			Validate: VRequired("Task name"),
			Parse:    ParseNonEmpty("Task name"),
			Assign:   func(h any, v any) { AssignString("Name", h, v) },
		},
		// tag (optional)
		{
			Label:   "Tag (optional)",
			Initial: "",
			Parse:   ParseOptString,
			Assign:  func(h any, v any) { Assign("Tag", h, v) },
		},
		// description (optional)
		{
			Label:   "Description (optional)",
			Initial: "",
			Parse:   ParseOptString,
			Assign:  func(h any, v any) { Assign("Description", h, v) },
		},
		// target (optional datetime → null.Time)
		{
			Label:    "Target date (YYYY-MM-DD, optional)",
			Initial:  "",
			Validate: VDateOptional(),
			Parse:    ParseOptDate,
			Assign:   func(h any, v any) { Assign("Target", h, v) },
		},
		// start (optional datetime → null.Time)
		{
			Label:    "Start date (YYYY-MM-DD, optional)",
			Initial:  "",
			Validate: VDateOptional(),
			Parse:    ParseOptDate,
			Assign:   func(h any, v any) { Assign("Start", h, v) },
		},
		// {
		//  Label:   "Archived (true/false, optional)",
		//  Initial: "",
		//  Parse:   ParseBool,               // returns bool
		//  Assign:  func(h any, v any) { Assign("Archived", h, null.BoolFrom(v.(bool))) },
		// },
	}

	RunFormWizard(fields, task)

	if err := task.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert task: %v", err)
	}
	fmt.Printf("Created task %d\n", task.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskEdit(_ *cobra.Command, args []string) {
	rawID := args[0]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid task ID %q: %v", rawID, err)
	}
	task, err := models.FindTask(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("couldn't find task %d: %v", idNum, err)
	}

	fields := []Field{
		// name (required)
		{
			Label:    "Task name",
			Initial:  task.Name,
			Validate: VRequired("Task name"),
			Parse:    ParseNonEmpty("Task name"),
			Assign:   func(h any, v any) { AssignString("Name", h, v) },
		},
		// tag (optional)
		{
			Label:   "Tag (optional)",
			Initial: task.Tag.String,
			Parse:   ParseOptString,
			Assign:  func(h any, v any) { Assign("Tag", h, v) },
		},
		// description (optional)
		{
			Label:   "Description (optional)",
			Initial: task.Description.String,
			Parse:   ParseOptString,
			Assign:  func(h any, v any) { Assign("Description", h, v) },
		},
		// target (optional datetime)
		{
			Label: "Target date (YYYY-MM-DD, optional)",
			Initial: func() string {
				if task.Target.Valid {
					return task.Target.Time.Format("2006-01-02")
				}
				return ""
			}(),
			Validate: VDateOptional(),
			Parse:    ParseOptDate,
			Assign:   func(h any, v any) { Assign("Target", h, v) },
		},
		// start (optional datetime)
		{
			Label: "Start date (YYYY-MM-DD, optional)",
			Initial: func() string {
				if task.Start.Valid {
					return task.Start.Time.Format("2006-01-02")
				}
				return ""
			}(),
			Validate: VDateOptional(),
			Parse:    ParseOptDate,
			Assign:   func(h any, v any) { Assign("Start", h, v) },
		},
		// {
		//  Label:   "Archived (true/false, optional)",
		//  Initial: strconv.FormatBool(task.Archived.Bool),
		//  Parse:   ParseBool,
		//  Assign:  func(h any, v any) { Assign("Archived", h, null.BoolFrom(v.(bool))) },
		// },
	}

	RunFormWizard(fields, task)

	if _, err := task.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update failed: %v", err)
	}
	fmt.Printf("Updated task %d\n", task.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
