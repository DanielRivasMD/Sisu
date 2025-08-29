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
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var milestoneCmd = &cobra.Command{
	Use:     "milestone",
	Short:   "Manage milestones",
	Long:    helpMilestone,
	Example: exampleMilestone,

	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var milestoneAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new milestone",
	Run:   runMilestoneAdd,
}

var milestoneEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Interactive TUI to edit milestone",
	Run:   runMilestoneEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(milestoneCmd)
	milestoneCmd.AddCommand(milestoneAddCmd, milestoneEditCmd)

	RegisterCrudSubcommands(milestoneCmd, "sisu.db", CrudModel[*models.Milestone]{
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
			ach = m.Achieved.Format("2006-01-02")

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
	})

}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runMilestoneAdd(_ *cobra.Command, args []string) {
	m := &models.Milestone{}

	fields := []Field{
		{
			Label:   "Task ID",
			Initial: "",
			Parse: func(s string) (any, error) {
				id, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid task ID: %w", err)
				}
				return id, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Task").
					SetInt(v.(int64))
			},
		},
		{
			Label:   "Type (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Type").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Value (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				if s == "" {
					return null.Int64{}, nil
				}
				v, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value: %w", err)
				}
				return null.Int64From(v), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Value").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Achieved date (YYYY-MM-DD, optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				if s == "" {
					return null.Time{}, nil
				}
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Achieved").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Message (optional)",
			Initial: "",
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Message").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, m)

	if err := m.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert milestone: %v", err)
	}
	fmt.Printf("Created milestone %d\n", m.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runMilestoneEdit(_ *cobra.Command, args []string) {
	raw := args[0]
	idNum, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		log.Fatalf("invalid milestone ID %q: %v", raw, err)
	}

	// FindMilestone expects null.Int64
	idNull := null.Int64From(idNum)
	m, err := models.FindMilestone(context.Background(), db.Conn, idNull)
	if err != nil {
		log.Fatalf("find milestone %d: %v", idNum, err)
	}

	fields := []Field{
		{
			Label:   "Task ID",
			Initial: strconv.FormatInt(m.Task, 10),
			Parse: func(s string) (any, error) {
				id, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid task ID: %w", err)
				}
				return id, nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Task").
					SetInt(v.(int64))
			},
		},
		{
			Label:   "Type (optional)",
			Initial: m.Type.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Type").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label: "Value (optional)",
			Initial: func() string {
				if m.Value.Valid {
					return strconv.FormatInt(m.Value.Int64, 10)
				}
				return ""
			}(),
			Parse: func(s string) (any, error) {
				if s == "" {
					return null.Int64{}, nil
				}
				v, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid value: %w", err)
				}
				return null.Int64From(v), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Value").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label: "Achieved date (YYYY-MM-DD, optional)",
			Initial: func() string {
				return m.Achieved.Format("2006-01-02")
			}(),
			Parse: func(s string) (any, error) {
				if s == "" {
					return null.Time{}, nil
				}
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Achieved").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Message (optional)",
			Initial: m.Message.String,
			Parse: func(s string) (any, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder any, v any) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Message").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, m)

	if _, err := m.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update milestone: %v", err)
	}
	fmt.Printf("Updated milestone %d\n", m.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
