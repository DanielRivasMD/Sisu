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

// sessionCmd is the parent for all "session" subcommands
var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage work sessions",

	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,
}

var sessionAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new session",
	Run:   runSessionAdd,
}

var sessionEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit an existing session",
	Args:  cobra.ExactArgs(1),
	Run:   runSessionEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionAddCmd, sessionEditCmd)

	RegisterCrudSubcommands(sessionCmd, "sisu.db", CrudModel[*models.Session]{
		Singular: "session",

		// ListFn: SELECT * FROM sessions ORDER BY id ASC
		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Session, error) {
			return models.Sessions(
				qm.OrderBy("id ASC"),
			).All(ctx, conn)
		},

		// Format: how each row shows up in "sisu session list"
		Format: func(s *models.Session) (int64, string) {
			// format the NOT NULL date
			date := s.Date.Time.Format("2006-01-02")

			// duration and score_feedback are nullable
			dur := ""
			if s.DurationMins.Valid {
				dur = strconv.FormatInt(s.DurationMins.Int64, 10)
			}
			score := ""
			if s.ScoreFeedback.Valid {
				score = strconv.FormatInt(s.ScoreFeedback.Int64, 10)
			}

			// notes is nullable text
			notes := s.Notes.String

			// task FK is a plain int64
			return s.ID.Int64, fmt.Sprintf(
				"task=%d date=%s dur=%s score=%s notes=%s",
				s.Task,
				date,
				dur,
				score,
				notes,
			)
		},

		// RemoveFn: DELETE FROM sessions WHERE id=?
		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			sess, err := models.FindSession(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = sess.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runSessionAdd launches a Bubble Tea form to gather session fields
func runSessionAdd(_ *cobra.Command, _ []string) {
	sess := &models.Session{}

	fields := []Field{
		{
			Label:   "Task ID",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			Assign: func(holder interface{}, v interface{}) {
				// TaskID is null.Int64
				reflect.ValueOf(holder).
					Elem().
					FieldByName("TaskID").
					Set(reflect.ValueOf(null.Int64From(v.(int64))))
			},
		},
		{
			Label:   "Session date (YYYY-MM-DD)",
			Initial: time.Now().Format("2006-01-02"),
			Parse: func(s string) (interface{}, error) {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Date").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Duration (minutes)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(n), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("DurationMinutes").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Score (1–5)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(n), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Score").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Notes (optional)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Notes").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, sess)

	if err := sess.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert session: %v", err)
	}
	fmt.Printf("Created session %d\n", sess.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runSessionEdit launches a pre‐seeded Bubble Tea form and then updates
func runSessionEdit(_ *cobra.Command, args []string) {
	idNum, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid session ID: %v", err)
	}
	ctx := db.Ctx()
	sess, err := models.FindSession(ctx, db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find session: %v", err)
	}

	fields := []Field{
		{
			Label:   "Task ID",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("TaskID").
					Set(reflect.ValueOf(null.Int64From(v.(int64))))
			},
		},
		{
			Label:   "Session date (YYYY-MM-DD)",
			Initial: time.Now().Format("2006-01-02"),
			Parse: func(s string) (interface{}, error) {
				t, err := time.Parse("2006-01-02", s)
				if err != nil {
					return nil, err
				}
				return null.TimeFrom(t), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Date").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Duration (minutes)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(n), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("DurationMinutes").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Score (1–5)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
				return null.Int64From(n), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Score").
					Set(reflect.ValueOf(v))
			},
		},
		{
			Label:   "Notes (optional)",
			Initial: sess.Notes.String,
			Parse: func(s string) (interface{}, error) {
				return null.StringFrom(s), nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Notes").
					Set(reflect.ValueOf(v))
			},
		},
	}

	RunFormWizard(fields, sess)

	if _, err := sess.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update session: %v", err)
	}
	fmt.Printf("Updated session %d\n", sess.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
