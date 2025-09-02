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
	"strings"
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

// sisu task archive [id ...]
// Archives one or more tasks by setting archived = true.
var taskArchiveCmd = &cobra.Command{
	Use:   "archive [id ...]",
	Short: "Archive tasks",
	Long:  helpTaskArchived,
	Args:  cobra.MinimumNArgs(1),
	Run:   runTaskArchive,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskAddCmd, taskEditCmd)
	taskCmd.AddCommand(taskArchiveCmd)

	RegisterCrudSubcommands(taskCmd, "sisu.db", CrudModel[*models.Task]{
		Singular: "task",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Task, error) {
			return models.Tasks(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(t *models.Task) (int64, string) {
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
			return t.ID.Int64, fmt.Sprintf("name=%s tag=%s target=%s start=%s archived=%v desc=%s",
				t.Name, tag, target, start, t.Archived.Bool, desc)
		},

		// pretty table
		TableHeaders: []string{"id", "name", "tag", "description", "start", "target", "archived"},
		TableRow: func(t *models.Task) []string {
			start := ""
			if t.Start.Valid {
				// Your sample shows full timestamp with offset; RFC3339 does that.
				start = t.Start.Time.Format(time.RFC3339)
			}
			target := ""
			if t.Target.Valid {
				target = t.Target.Time.Format(time.RFC3339)
			}
			archived := "0"
			if t.Archived.Bool {
				archived = "1"
			}
			return []string{
				strconv.FormatInt(t.ID.Int64, 10),
				t.Name,
				t.Tag.String,
				t.Description.String,
				start,
				target,
				archived,
			}
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			task, err := models.FindTask(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = task.Delete(ctx, conn)
			return err
		},
		HintFn: func(t *models.Task) string {
			start, target := "", ""
			if t.Start.Valid {
				start = t.Start.Time.Format("2006-01-02")
			}
			if t.Target.Valid {
				target = t.Target.Time.Format("2006-01-02")
			}
			archived := "0"
			if t.Archived.Bool {
				archived = "1"
			}
			return fmt.Sprintf("name, %s tag, %s start, %s target, %s archived, %s",
				t.Name, t.Tag.String, start, target, archived)
		},
	})

	AttachEditCompletion(taskEditCmd,
		func(ctx context.Context, conn *sql.DB) ([]*models.Task, error) {
			return models.Tasks(qm.OrderBy("id ASC")).All(ctx, conn)
		},
		func(t *models.Task) (int64, string) { // format fallback (id + simple)
			return t.ID.Int64, t.Name
		},
		func(t *models.Task) string { // rich hint
			start, target := "", ""
			if t.Start.Valid {
				start = t.Start.Time.Format("2006-01-02")
			}
			if t.Target.Valid {
				target = t.Target.Time.Format("2006-01-02")
			}
			archived := "0"
			if t.Archived.Bool {
				archived = "1"
			}
			return fmt.Sprintf("name, %s tag, %s start, %s target, %s archived, %s",
				t.Name, t.Tag.String, start, target, archived)
		},
	)

}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskAdd(_ *cobra.Command, _ []string) {
	task := &models.Task{}

	// capture the chosen Start date (nullable) to drive Target default & seeding
	var startPicked null.Time

	// convenience for showing initial suggestion before user edits Start
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	profileChoice := "" // "default" or "custom"

	fields := []Field{
		// profile choice
		{
			Label:    "Profile (default/custom)",
			Initial:  "default",
			Validate: VRequired("Profile"),
			Parse: func(s string) (any, error) {
				v := strings.ToLower(strings.TrimSpace(s))
				switch v {
				case "default", "custom":
					return v, nil
				default:
					return nil, fmt.Errorf("must be 'default' or 'custom'")
				}
			},
			Assign: func(h any, v any) {
				profileChoice = v.(string)
			},
		},

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

		// start (optional datetime → null.Time) defaults to today
		{
			Label:    "Start date (YYYY-MM-DD, optional)",
			Initial:  todayStr,
			Validate: VDateOptional(),
			Parse:    ParseOptDate,
			Assign: func(h any, v any) {
				// assign to model
				Assign("Start", h, v)
				// keep a copy to compute Target & seed dates
				startPicked = v.(null.Time)
			},
		},

		// target (optional datetime → null.Time) defaults to today + 100 days
		{
			Label:    "Target date (YYYY-MM-DD, optional) Default calculated to 100 days from Start date",
			Initial:  "",
			Validate: VDateOptional(),
			Parse: func(s string) (any, error) {
				s = strings.TrimSpace(s)
				if s != "" {
					return ParseOptDate(s)
				}
				// user left blank → compute default from Start (or today if Start is empty)
				base := today
				if startPicked.Valid {
					base = startPicked.Time
				}
				return null.TimeFrom(base.AddDate(0, 0, 100)), nil
			},
			Assign: func(h any, v any) { Assign("Target", h, v) },
		},
	}

	RunFormWizard(fields, task)

	// Insert the task
	if err := task.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert task: %v", err)
	}
	fmt.Printf("Created task %d\n", task.ID.Int64)

	base := today
	if startPicked.Valid {
		base = startPicked.Time
	}

	// If profileChoice is "default", seed related tables
	if profileChoice == "default" {
		if err := seedDefaultTaskProfileWithBase(db.Ctx(), db.Conn, task.ID.Int64, base); err != nil {
			log.Fatalf("seed default profile: %v", err)
		}
		fmt.Println("Applied default profile (milestones, reviews, coach).")
	}
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
	}

	RunFormWizard(fields, task)

	if _, err := task.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update failed: %v", err)
	}
	fmt.Printf("Updated task %d\n", task.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// seedDefaultTaskProfile seeds milestones, reviews, and coach defaults for a task.
func seedDefaultTaskProfileWithBase(ctx context.Context, exec boil.ContextExecutor, taskID int64, base time.Time) error {
	// Helpers for dates
	addDays := func(n int) null.Time { return null.TimeFrom(base.AddDate(0, 0, n)) }

	// 1) Milestones (three entries)
	milestones := []*models.Milestone{
		{
			Task:    taskID,
			Type:    null.StringFrom("courage"),
			Value:   null.Int64From(2),
			Done:    addDays(25),
			Message: null.StringFrom("face difficulties with resolve"),
		},
		{
			Task:    taskID,
			Type:    null.StringFrom("determination"),
			Value:   null.Int64From(3),
			Done:    addDays(50),
			Message: null.StringFrom("continue despite challenges"),
		},
		{
			Task:    taskID,
			Type:    null.StringFrom("perseverance"),
			Value:   null.Int64From(4),
			Done:    addDays(75),
			Message: null.StringFrom("stay committed to the goal"),
		},
	}
	for _, m := range milestones {
		if err := m.Insert(ctx, exec, boil.Infer()); err != nil {
			return fmt.Errorf("insert milestone: %w", err)
		}
	}

	// 2) Reviews: seven entries with weeks 2,4,6,8,10,12,14 and ordinal summaries
	weeks := []int64{2, 4, 6, 8, 10, 12, 14}
	ordinals := []string{"first", "second", "third", "fourth", "fifth", "sixth", "seventh"}
	for i, w := range weeks {
		r := &models.Review{
			Task:    taskID,
			Week:    null.Int64From(w),
			Summary: null.StringFrom(ordinals[i] + ": "),
		}
		if err := r.Insert(ctx, exec, boil.Infer()); err != nil {
			return fmt.Errorf("insert review (week %d): %w", w, err)
		}
	}

	// 3) Coach: three entries (date optional -> leave NULL)
	coaches := []*models.Coach{
		{Trigger: "forgotten", Content: "forgotten"},
		{Trigger: "low performance", Content: "low performance"},
		{Trigger: "superb", Content: "superb"},
	}
	for _, c := range coaches {
		if err := c.Insert(ctx, exec, boil.Infer()); err != nil {
			return fmt.Errorf("insert coach: %w", err)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runTaskArchive(cmd *cobra.Command, args []string) {
	ctx := db.Ctx()

	// Iterate all provided IDs
	for _, raw := range args {
		idNum, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			log.Fatalf("invalid task ID %q: %v", raw, err)
		}

		// Load the task
		t, err := models.FindTask(ctx, db.Conn, null.Int64From(idNum))
		if err != nil {
			log.Fatalf("find task %d: %v", idNum, err)
		}

		t.Archived = null.BoolFrom(true)

		if _, err := t.Update(ctx, db.Conn, boil.Whitelist("archived")); err != nil {
			log.Fatalf("archive task %d: %v", idNum, err)
		}
		fmt.Printf("Archived task %d\n", idNum)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
