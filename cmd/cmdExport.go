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
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	exportAll bool

	exportCmd = &cobra.Command{
		Use:               "export [tables...]",
		Short:             "Export one or more tables to CSV files",
		Long:              helpExport,
		Example:           exampleExport,
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,

		Args: cobra.ArbitraryArgs,

		Run: runExport,
	}
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
		cmd.Help()
		return
	}

	for _, table := range args {
		switch table {
		case "tasks":
			if err := exportTasks(db.Ctx(), db.Conn); err != nil {
				log.Fatalf("export tasks: %v", err)
			}
		case "sessions":
			if err := exportSessions(db.Ctx(), db.Conn); err != nil {
				log.Fatalf("export sessions: %v", err)
			}
		case "milestones":
			if err := exportMilestones(db.Ctx(), db.Conn); err != nil {
				log.Fatalf("export milestones: %v", err)
			}
		case "reviews":
			if err := exportReviews(db.Ctx(), db.Conn); err != nil {
				log.Fatalf("export reviews: %v", err)
			}
		case "coach":
			if err := exportCoach(db.Ctx(), db.Conn); err != nil {
				log.Fatalf("export coach: %v", err)
			}
		case "calendar":
			if err := exportCalendar(db.Ctx(), db.Conn); err != nil {
				log.Fatalf("export calendar: %v", err)
			}
		default:
			log.Fatalf("unknown table %q", table)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// --- exportTasks ---
func exportTasks(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Tasks(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query tasks: %w", err)
	}

	file, err := os.Create("tasks.csv")
	if err != nil {
		return fmt.Errorf("create tasks.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"id", "name", "tag", "description", "date_target", "date_start", "archived"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, t := range rows {
		dt := ""
		if t.DateTarget.Valid {
			dt = t.DateTarget.Time.Format(time.RFC3339)
		}
		ds := ""
		if t.DateStart.Valid {
			ds = t.DateStart.Time.Format(time.RFC3339)
		}
		rec := []string{
			strconv.FormatInt(t.ID.Int64, 10),
			t.Name,
			t.Tag.String,
			t.Description.String,
			dt,
			ds,
			strconv.FormatBool(t.Archived.Bool),
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	fmt.Println("exported tasks.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// --- exportSessions ---
func exportSessions(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Sessions(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query sessions: %w", err)
	}

	file, err := os.Create("sessions.csv")
	if err != nil {
		return fmt.Errorf("create sessions.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"id", "task", "date", "duration_mins", "score_feedback", "notes"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, s := range rows {
		date := s.Date.Format("2006-01-02")
		dur := ""
		if s.DurationMins.Valid {
			dur = strconv.FormatInt(s.DurationMins.Int64, 10)
		}
		score := ""
		if s.ScoreFeedback.Valid {
			score = strconv.FormatInt(s.ScoreFeedback.Int64, 10)
		}
		rec := []string{
			strconv.FormatInt(s.ID.Int64, 10),
			strconv.FormatInt(s.Task, 10),
			date,
			dur,
			score,
			s.Notes.String,
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	fmt.Println("exported sessions.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// --- exportMilestones ---
func exportMilestones(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Milestones(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query milestones: %w", err)
	}

	file, err := os.Create("milestones.csv")
	if err != nil {
		return fmt.Errorf("create milestones.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"id", "task", "type", "value", "achieved", "message"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, m := range rows {
		val := ""
		if m.Value.Valid {
			val = strconv.FormatInt(m.Value.Int64, 10)
		}
		ach := ""
		if m.Achieved.Valid {
			ach = m.Achieved.Time.Format("2006-01-02")
		}
		rec := []string{
			strconv.FormatInt(m.ID.Int64, 10),
			strconv.FormatInt(m.Task, 10),
			m.Type.String,
			val,
			ach,
			m.Message.String,
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	fmt.Println("exported milestones.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// --- exportReviews ---
func exportReviews(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Reviews(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query reviews: %w", err)
	}

	file, err := os.Create("reviews.csv")
	if err != nil {
		return fmt.Errorf("create reviews.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"id", "task", "week", "summary"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		wk := ""
		if r.Week.Valid {
			wk = strconv.FormatInt(r.Week.Int64, 10)
		}
		rec := []string{
			strconv.FormatInt(r.ID.Int64, 10),
			strconv.FormatInt(r.Task, 10),
			wk,
			r.Summary.String,
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	fmt.Println("exported reviews.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// --- exportCoach ---
func exportCoach(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Coaches(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query coach: %w", err)
	}

	file, err := os.Create("coach.csv")
	if err != nil {
		return fmt.Errorf("create coach.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"id", "trigger", "content", "date"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, c := range rows {
		date := ""
		if c.Date.Valid {
			date = c.Date.Time.Format("2006-01-02")
		}
		rec := []string{
			strconv.FormatInt(c.ID.Int64, 10),
			c.Trigger,
			c.Content,
			date,
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	fmt.Println("exported coach.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// --- exportCalendar ---
func exportCalendar(ctx context.Context, conn *sql.DB) error {
	rows, err := models.Calendars(qm.OrderBy("id ASC")).All(ctx, conn)
	if err != nil {
		return fmt.Errorf("query calendar: %w", err)
	}

	file, err := os.Create("calendar.csv")
	if err != nil {
		return fmt.Errorf("create calendar.csv: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"id", "date", "note"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, c := range rows {
		rec := []string{
			strconv.FormatInt(c.ID.Int64, 10),
			c.Date.Format("2006-01-02"),
			c.Note,
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	fmt.Println("exported calendar.csv")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
