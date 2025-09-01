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

// Schema (sessions):
//   id INTEGER PK
//   task INTEGER NOT NULL  → int64 (required)
//   date DATE              → null.Time (optional)
//   mins INTEGER           → null.Int64 (optional)
//   feedback INTEGER       → null.Int64 (optional)
//   notes TEXT             → null.String (optional)

var sessionCmd = &cobra.Command{
	Use:               "session",
	Short:             "Manage work sessions",
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

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Session, error) {
			return models.Sessions(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(s *models.Session) (int64, string) {
			date := ""
			if s.Date.Valid {
				date = s.Date.Time.Format("2006-01-02")
			}
			mins := ""
			if s.Mins.Valid {
				mins = strconv.FormatInt(s.Mins.Int64, 10)
			}
			fb := ""
			if s.Feedback.Valid {
				fb = strconv.FormatInt(s.Feedback.Int64, 10)
			}
			notes := s.Notes.String

			return s.ID.Int64, fmt.Sprintf(
				"task=%d date=%s mins=%s feedback=%s notes=%s",
				s.Task, date, mins, fb, notes,
			)
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			row, err := models.FindSession(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = row.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSessionAdd(_ *cobra.Command, _ []string) {
	sess := &models.Session{}

	fields := []Field{
		// required FK
		FInt("Task ID", "Task", ""),
		// optional fields matching schema
		FOptDate("Session date (YYYY-MM-DD, optional)", "Date", ""),
		FOptInt("Duration (minutes, optional)", "Mins", ""),
		FOptInt("Score (1–5, optional)", "Feedback", ""),
		FOptString("Notes (optional)", "Notes", ""),
	}

	RunFormWizard(fields, sess)

	if err := sess.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert session: %v", err)
	}
	fmt.Printf("Created session %d\n", sess.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

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
		FInt("Task ID", "Task", strconv.FormatInt(sess.Task, 10)),
		FOptDate("Session date (YYYY-MM-DD, optional)", "Date",
			func() string {
				if sess.Date.Valid {
					return sess.Date.Time.Format("2006-01-02")
				}
				return ""
			}(),
		),
		FOptInt("Duration (minutes, optional)", "Mins",
			func() string {
				if sess.Mins.Valid {
					return strconv.FormatInt(sess.Mins.Int64, 10)
				}
				return ""
			}(),
		),
		FOptInt("Score (1–5, optional)", "Feedback",
			func() string {
				if sess.Feedback.Valid {
					return strconv.FormatInt(sess.Feedback.Int64, 10)
				}
				return ""
			}(),
		),
		FOptString("Notes (optional)", "Notes", sess.Notes.String),
	}

	RunFormWizard(fields, sess)

	if _, err := sess.Update(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("update session: %v", err)
	}
	fmt.Printf("Updated session %d\n", sess.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
