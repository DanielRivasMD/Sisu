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
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var trackCmd = &cobra.Command{
	Use:     "track",
	Short:   "",
	Long:    helpTrack,
	Example: exampleTrack,

	PreRun: preRunTrack,
	Run: runTrack,
	PostRun: postRunTrack,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(trackCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpTrack = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

var exampleTrack = formatExample(
	"",
	[]string{"track"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func preRunTrack(cmd *cobra.Command, args []string) {
	conn, err := db.InitDB(dbPath)
	if err != nil {
		fmt.Println("database initialization failed: %w", err)
	}
	_ = conn
}

func runTrack(cmd *cobra.Command, args []string) {
	// 1) Grab the shared *sql.DB
	conn := db.Conn
	if conn == nil {
		fmt.Fprintln(os.Stderr, "database not initialized")
		os.Exit(1)
	}

	// 2) Select or create a task
	taskID, err := promptSelectOrCreateTask(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "task selection error: %v\n", err)
		os.Exit(1)
	}

	// 3) Prompt for session details
	session, err := promptSessionDetails()
	if err != nil {
		fmt.Fprintf(os.Stderr, "input error: %v\n", err)
		os.Exit(1)
	}

	// 4) Insert into sessions
	_, err = conn.Exec(
		`INSERT INTO sessions(task_id, date, duration_minutes, score, notes)
     VALUES (?, ?, ?, ?, ?)`,
		taskID,
		session.Date.Format("2006-01-02"),
		session.Duration,
		session.Score,
		session.Notes,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to insert session: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session recorded!")
}

func postRunTrack(cmd *cobra.Command, args []string) {
	if db.Conn != nil {
		if err := db.Conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing database: %v\n", err)
		}
		db.Conn = nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// promptSelectOrCreateTask shows existing tasks, or lets you type a new one.
func promptSelectOrCreateTask(db *sql.DB) (int64, error) {
	// fetch existing tasks
	rows, err := db.Query(`SELECT id, name FROM tasks WHERE archived = 0 ORDER BY name`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var ids []int64
	var names []string
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return 0, err
		}
		ids = append(ids, id)
		names = append(names, name)
	}

	// add option to create new
	names = append(names, "[Create new task]")

	var choice string
	prompt := &survey.Select{
		Message: "Which task?",
		Options: names,
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return 0, err
	}

	// if choosing to create
	if choice == "[Create new task]" {
		var newName string
		if err := survey.AskOne(&survey.Input{
			Message: "Task name:",
		}, &newName, survey.WithValidator(survey.Required)); err != nil {
			return 0, err
		}

		res, err := db.Exec(
			`INSERT INTO tasks(name, created_at) VALUES (?, ?)`,
			newName,
			time.Now().Format(time.RFC3339),
		)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}

	// otherwise return selected id
	for i, n := range names {
		if n == choice {
			return ids[i], nil
		}
	}
	return 0, fmt.Errorf("invalid task choice")
}

// sessionData holds the answers for a track entry
type sessionData struct {
	Date     time.Time
	Duration int
	Score    int
	Notes    string
}

func promptSessionDetails() (*sessionData, error) {
	var sd sessionData
	// Date
	dateStr := time.Now().Format("2006-01-02")
	if err := survey.AskOne(&survey.Input{
		Message: "Date (YYYY-MM-DD):",
		Default: dateStr,
	}, &dateStr, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	sd.Date = d

	// Duration
	if err := survey.AskOne(&survey.Input{
		Message: "Duration (minutes):",
	}, &sd.Duration, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	// Score
	if err := survey.AskOne(&survey.Input{
		Message: "Score (1–5):",
	}, &sd.Score, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	// Notes
	survey.AskOne(&survey.Input{
		Message: "Notes (optional):",
	}, &sd.Notes)

	return &sd, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
