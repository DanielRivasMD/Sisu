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
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var exportCmd = &cobra.Command{
	Use:               "export [tables...]",
	Short:             "Export one or more tables to CSV files",
	Long:              helpExport,
	Example:           exampleExport,
	PersistentPreRun:  persistentPreRun,
	PersistentPostRun: persistentPostRun,

	Args: cobra.ArbitraryArgs,

	Run: runExport,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	exportAll bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all tables")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runExport(cmd *cobra.Command, args []string) {
	if exportAll {
		args = []string{"tasks", "sessions", "milestones", "reviews", "coach", "calendar"}
	}
	if len(args) == 0 {
		horus.CheckErr(cmd.Help())
		return
	}

	ctx := db.Ctx()
	exec := boil.ContextExecutor(db.Conn)

	for _, table := range args {
		switch table {
		case "tasks":
			horus.CheckErr(exportTable(
				ctx, exec,
				"tasks.csv",
				[]string{"id", "name", "tag", "description", "date_target", "date_start", "archived"},
				models.Tasks(qm.OrderBy("id ASC")).All,
				func(t *models.Task) []string {
					dt, ds := "", ""
					if t.DateTarget.Valid {
						dt = t.DateTarget.Time.Format(time.RFC3339)
					}
					if t.DateStart.Valid {
						ds = t.DateStart.Time.Format(time.RFC3339)
					}
					return []string{
						strconv.FormatInt(t.ID.Int64, 10),
						t.Name,
						t.Tag.String,
						t.Description.String,
						dt,
						ds,
						strconv.FormatBool(t.Archived.Bool),
					}
				},
			))

		case "sessions":
			horus.CheckErr(exportTable(
				ctx, exec,
				"sessions.csv",
				[]string{"id", "task", "date", "duration_mins", "score_feedback", "notes"},
				models.Sessions(qm.OrderBy("id ASC")).All,
				func(s *models.Session) []string {
					dur, score := "", ""
					if s.DurationMins.Valid {
						dur = strconv.FormatInt(s.DurationMins.Int64, 10)
					}
					if s.ScoreFeedback.Valid {
						score = strconv.FormatInt(s.ScoreFeedback.Int64, 10)
					}
					return []string{
						strconv.FormatInt(s.ID.Int64, 10),
						strconv.FormatInt(s.Task, 10),
						s.Date.Format("2006-01-02"),
						dur,
						score,
						s.Notes.String,
					}
				},
			))

		case "milestones":
			horus.CheckErr(exportTable(
				ctx, exec,
				"milestones.csv",
				[]string{"id", "task", "type", "value", "achieved", "message"},
				models.Milestones(qm.OrderBy("id ASC")).All,
				func(m *models.Milestone) []string {
					val, ach := "", ""
					if m.Value.Valid {
						val = strconv.FormatInt(m.Value.Int64, 10)
					}
					if m.Achieved.Valid {
						ach = m.Achieved.Time.Format("2006-01-02")
					}
					return []string{
						strconv.FormatInt(m.ID.Int64, 10),
						strconv.FormatInt(m.Task, 10),
						m.Type.String,
						val,
						ach,
						m.Message.String,
					}
				},
			))

		case "reviews":
			horus.CheckErr(exportTable(
				ctx, exec,
				"reviews.csv",
				[]string{"id", "task", "week", "summary"},
				models.Reviews(qm.OrderBy("id ASC")).All,
				func(r *models.Review) []string {
					wk := ""
					if r.Week.Valid {
						wk = strconv.FormatInt(r.Week.Int64, 10)
					}
					return []string{
						strconv.FormatInt(r.ID.Int64, 10),
						strconv.FormatInt(r.Task, 10),
						wk,
						r.Summary.String,
					}
				},
			))

		case "coach":
			horus.CheckErr(exportTable(
				ctx, exec,
				"coach.csv",
				[]string{"id", "trigger", "content", "date"},
				models.Coaches(qm.OrderBy("id ASC")).All,
				func(c *models.Coach) []string {
					date := ""
					if c.Date.Valid {
						date = c.Date.Time.Format("2006-01-02")
					}
					return []string{
						strconv.FormatInt(c.ID.Int64, 10),
						c.Trigger,
						c.Content,
						date,
					}
				},
			))

		case "calendar":
			horus.CheckErr(exportTable(
				ctx, exec,
				"calendar.csv",
				[]string{"id", "date", "note"},
				models.Calendars(qm.OrderBy("id ASC")).All,
				func(c *models.Calendar) []string {
					return []string{
						strconv.FormatInt(c.ID.Int64, 10),
						c.Date.Format("2006-01-02"),
						c.Note,
					}
				},
			))

		default:
			log.Fatalf("unknown table %q", table)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// exportTable exports any SQLBoiler slice type S whose elements are T.
//
//	S must be a defined slice type with underlying []T (e.g. models.CoachSlice ~ []*models.Coach)
//	fetchFn must match S’s All() signature.
//	formatFn turns each T → []string for CSV writing.
func exportTable[T any, S ~[]T](
	ctx context.Context,
	exec boil.ContextExecutor,
	fileName string,
	header []string,
	fetchFn func(context.Context, boil.ContextExecutor) (S, error),
	formatFn func(item T) []string,
) error {
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("create %s: %w", fileName, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(header); err != nil {
		return fmt.Errorf("write header for %s: %w", fileName, err)
	}

	rows, err := fetchFn(ctx, exec)
	if err != nil {
		return fmt.Errorf("query %s: %w", fileName, err)
	}

	for _, r := range rows {
		if err := w.Write(formatFn(r)); err != nil {
			return fmt.Errorf("write record for %s: %w", fileName, err)
		}
	}

	fmt.Println("exported", fileName)
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
