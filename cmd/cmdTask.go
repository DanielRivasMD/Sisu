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

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:    helpTask,
	Example: exampleTask,

	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var taskAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new task",
	Run:   runTaskAdd,
}

var taskEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Interactive TUI to edit task",
	Run:   runTaskEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskAddCmd, taskEditCmd)

	RegisterCrudSubcommands(taskCmd, "sisu.db", CrudModel[*models.Task]{
		Singular: "task",

		ListFn: func(ctx context.Context, db *sql.DB) ([]*models.Task, error) {
			return models.Tasks(qm.OrderBy("id ASC")).All(ctx, db)
		},

		Format: func(t *models.Task) (int64, string) {
			return t.ID.Int64, fmt.Sprintf("%s (archived=%v)", t.Name, t.Archived.Bool)
		},

		RemoveFn: func(ctx context.Context, db *sql.DB, id int64) error {
			task, err := models.FindTask(ctx, db, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = task.Delete(ctx, db)
			return err
		},
	})

}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskAdd(_ *cobra.Command, args []string) {
	task := &models.Task{}

	fields := []Field{
		{
			Label:   "Task name",
			Initial: "",
			Parse: func(s string) (any, error) {
				if s == "" {
					return nil, fmt.Errorf("name cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Name").
					SetString(v.(string))
			},
		},
		{
			Label:   "Description (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Description").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Target date (YYYY-MM-DD)",
			Initial: time.Now().Format("2006-01-02"),
			Parse: func(s string) (any, error) {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("DateTarget").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, task)

	// Persist via your db package
	if err := task.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert task: %v", err)
	}
	fmt.Printf("Created task %d\n", task.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskEdit(_ *cobra.Command, args []string) {
	// 1) parse the CLI‐arg into an int64
	rawID := args[0]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid task ID %q: %v", rawID, err)
	}

	// 2) wrap into a null.Int64
	id := null.Int64From(idNum)

	// 3) call FindTask with a null.Int64
	task, err := models.FindTask(context.Background(), db.Conn, id)
	if err != nil {
		log.Fatalf("couldn't find task %d: %v", idNum, err)
	}

	// 4) seed and run the form wizard just like Add
	fields := []Field{
		{
			Label:   "Task name",
			Initial: task.Name,
			Parse: func(s string) (any, error) {
				if s == "" {
					return nil, fmt.Errorf("name cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("Name").
					SetString(v.(string))
			},
		},
		{
			Label:   "Description (optional)",
			Initial: task.Description.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("Description").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Target date (YYYY-MM-DD)",
			Initial: task.DateTarget.Time.Format("2006-01-02"),
			Parse: func(s string) (any, error) {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).Elem().FieldByName("DateTarget").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, task)

	// 5) persist with Update
	if _, err := task.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update failed: %v", err)
	}
	fmt.Printf("Updated task %d\n", task.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
