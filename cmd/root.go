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
	"strconv"
	"strings"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"

	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     "sisu",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	verbose bool
)

var (
	dbPath string // populated by the --db flag
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostics")

	rootCmd.PersistentFlags().
		StringVar(&dbPath, "db", "sisu.db", "path to sqlite database")

		// rootCmd.AddCommand(NewCrudCmd("sisu.db", CrudModel[*models.Task]{
		// 	Singular: "task",

		// 	ListFn: func(ctx context.Context, db *sql.DB) ([]*models.Task, error) {
		// 		return models.Tasks(qm.OrderBy("id ASC")).All(ctx, db)
		// 	},

		// 	Format: func(t *models.Task) (int64, string) {
		// 		return t.ID.Int64, fmt.Sprintf("%s (archived=%v)", t.Name, t.Archived.Bool)
		// 	},

		// 	AddFn: func(ctx context.Context, db *sql.DB, args []string) (int64, error) {
		// 		name := args[0]
		// 		t := &models.Task{
		// 			Name:        name,
		// 			Description: null.StringFrom(""),        // or args[1]
		// 			DateTarget:  null.TimeFrom(time.Time{}), // or parse args[2]
		// 			Archived:    null.BoolFrom(false),
		// 		}
		// 		if err := t.Insert(ctx, db, boil.Infer()); err != nil {
		// 			return 0, err
		// 		}
		// 		return t.ID.Int64, nil
		// 	},

		// 	RemoveFn: func(ctx context.Context, db *sql.DB, id int64) error {
		// 		task, err := models.FindTask(ctx, db, null.Int64From(id))
		// 		if err != nil {
		// 			return err
		// 		}
		// 		_, err = task.Delete(ctx, db)
		// 		return err
		// 	},

		// 	EditFn: func(ctx context.Context, db *sql.DB, id int64, args []string) error {
		// 		task, err := models.FindTask(ctx, db, null.Int64From(id))
		// 		if err != nil {
		// 			return err
		// 		}
		// 		task.Description = null.StringFrom(args[0])
		// 		_, err = task.Update(ctx, db, boil.Whitelist("description"))
		// 		return err
		// 	},
		// }))

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

			// 3. AddFn parses args and INSERTs a new row
			//    Usage: sisu milestone add <taskID> <type> <value> <achieved(YYYY-MM-DD)> <message>
			AddFn: func(ctx context.Context, db *sql.DB, args []string) (int64, error) {
				// parse taskID into a plain int64
				taskID, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("invalid task id: %w", err)
				}

				// parse value into a null.Int64
				val, err := strconv.ParseInt(args[2], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("invalid value: %w", err)
				}

				// optional achieved date
				var ach null.Time
				if d := strings.TrimSpace(args[3]); d != "" {
					if t, err := time.Parse("2006-01-02", d); err == nil {
						ach = null.TimeFrom(t)
					}
				}

				m := &models.Milestone{
					// plain int64
					Task: taskID,

					// nullable wrappers
					Type:     null.StringFrom(args[1]),
					Value:    null.Int64From(val),
					Achieved: ach,
					Message:  null.StringFrom(args[4]),
				}

				if err := m.Insert(ctx, db, boil.Infer()); err != nil {
					return 0, err
				}
				return m.ID.Int64, nil
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

			// 5. EditFn updates the message
			//    Usage: sisu milestone edit <id> <new message>
			EditFn: func(ctx context.Context, db *sql.DB, id int64, args []string) error {
				m, err := models.FindMilestone(ctx, db, null.Int64From(id))
				if err != nil {
					return err
				}

				m.Message = null.StringFrom(args[0])
				_, err = m.Update(ctx, db, boil.Whitelist("message"))
				return err
			},
		}),
	)

}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)
var exampleRoot = formatExample(
	"sisu",
	[]string{"help"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// formatHelp produces the “help” header + description.
//
//	author: name, e.g. "Daniel Rivas"
//	email:  email, e.g. "danielrivasmd@gmail.com"
//	desc:   the multi‐line description, "\n"-separated.
func formatHelp(author, email, desc string) string {
	header := chalk.Bold.TextStyle(
		chalk.Green.Color(author+" "),
	) +
		chalk.Dim.TextStyle(
			chalk.Italic.TextStyle("<"+email+">"),
		)

	// prefix two newlines to your desc, chalk it cyan + dim it
	body := "\n\n" + desc
	return header + chalk.Dim.TextStyle(chalk.Cyan.Color(body))
}

// formatExample builds a multi‐line example block
// each usage is a slice of “tokens”: [ command, flagOrArg, flagOrArg, ... ].
//
//	app:    your binary name, e.g. "lilith"
//	usages: one or more usages—each becomes its own line.
func formatExample(app string, usages ...[]string) string {
	var b strings.Builder

	for i, usage := range usages {
		if len(usage) == 0 {
			continue
		}

		// first token is the subcommand
		b.WriteString(
			chalk.White.Color(app) + " " +
				chalk.White.Color(chalk.Bold.TextStyle(usage[0])),
		)

		// remaining tokens are either flags (--foo) or args
		for _, tok := range usage[1:] {
			switch {
			case strings.HasPrefix(tok, "--"):
				b.WriteString(" " + chalk.Italic.TextStyle(chalk.White.Color(tok)))
			default:
				b.WriteString(" " + chalk.Dim.TextStyle(chalk.Italic.TextStyle(tok)))
			}
		}

		if i < len(usages)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
