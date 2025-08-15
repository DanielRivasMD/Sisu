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
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var trackCmd = &cobra.Command{
	Use:     "track",
	Short:   "",
	Long:    helpTrack,
	Example: exampleTrack,

	PreRun:  preRunTrack,
	Run:     runTrack,
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
	// ensure DB is initialized
	conn := db.Conn
	if conn == nil {
		log.Fatalln("database not initialized")
	}

	// load tasks
	rows, err := conn.Query(`SELECT id, name FROM tasks WHERE archived = 0 ORDER BY name`)
	if err != nil {
		log.Fatalf("fetch tasks: %v", err)
	}
	defer rows.Close()

	var tasks []taskItem
	for rows.Next() {
		var t taskItem
		if err := rows.Scan(&t.id, &t.name); err != nil {
			log.Fatalf("scan task: %v", err)
		}
		tasks = append(tasks, t)
	}

	// start TUI
	m := initialModel(tasks)
	p := tea.NewProgram(m)
	finalModel, err := p.StartReturningModel()
	if err != nil {
		log.Fatalf("starting TUI: %v", err)
	}
	m = finalModel.(model)

	// if new task, insert it
	if m.newTaskName != "" {
		res, err := conn.Exec(
			`INSERT INTO tasks(name, created_at) VALUES (?, ?)`,
			m.newTaskName,
			time.Now().Format("2006-01-02"),
		)
		if err != nil {
			log.Fatalf("insert new task: %v", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			log.Fatalf("lastInsertId: %v", err)
		}
		m.selectedID = id
	}

	// final insert into sessions
	_, err = conn.Exec(
		`INSERT INTO sessions(task_id, date, duration_minutes, score, notes)
       VALUES (?, ?, ?, ?, ?)`,
		m.selectedID,
		m.session.Date.Format("2006-01-02"),
		m.session.Duration,
		m.session.Score,
		m.session.Notes,
	)
	if err != nil {
		log.Fatalf("insert session: %v", err)
	}

	fmt.Println("\n✅ Session recorded!")
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

var (
	// configure list dimensions
	listWidth  = 40
	listHeight = 10
)

// sessionData holds the answers for a track entry
type sessionData struct {
	Date     time.Time
	Duration int
	Score    int
	Notes    string
}

type taskItem struct {
	id   int64
	name string
}

func (t taskItem) Title() string       { return t.name }
func (t taskItem) Description() string { return "" }
func (t taskItem) FilterValue() string { return t.name }

// stage constants
const (
	stageSelectTask = iota
	stageNewTaskName
	stageDateInput
	stageDurationInput
	stageScoreInput
	stageNotesInput
	stageDone
)

// model holds all TUI state
type model struct {
	stage       int
	list        list.Model
	textInput   textinput.Model
	tasks       []taskItem
	selectedID  int64
	newTaskName string
	session     sessionData
}

// initialModel builds the first stage (task list)
func initialModel(tasks []taskItem) model {
	items := make([]list.Item, len(tasks)+1)
	for i, t := range tasks {
		items[i] = t
	}
	// the last item is "Create new task"
	items[len(tasks)] = taskItem{id: 0, name: "[Create new task]"}

	l := list.New(items, list.NewDefaultDelegate(), listWidth, listHeight)
	l.Title = "Select a task:"
	l.SetShowHelp(false)

	// prepare a blank textinput (we’ll configure when we need it)
	ti := textinput.New()
	ti.CharLimit = 64
	ti.Width = listWidth

	return model{
		stage:     stageSelectTask,
		list:      l,
		textInput: ti,
		tasks:     tasks,
	}
}

// Init implements tea.Model
func (m model) Init() tea.Cmd {
	return nil
}

// Update drives the TUI state machine
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.stage {
	case stageSelectTask:
		// update the list
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				sel := m.list.SelectedItem().(taskItem)
				if sel.id == 0 {
					// new task path
					m.stage = stageNewTaskName
					m.textInput.Placeholder = "New task name"
					m.textInput.SetValue("")
					m.textInput.Focus()
				} else {
					// existing task selected
					m.selectedID = sel.id
					m.stage = stageDateInput
					m.textInput.Placeholder = "YYYY-MM-DD"
					m.textInput.SetValue(time.Now().Format("2006-01-02"))
					m.textInput.Focus()
				}
				return m, nil
			}
		}
		return m, cmd

	case stageNewTaskName:
		// text input for new task name
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			m.newTaskName = m.textInput.Value()
			m.stage = stageDateInput
			m.textInput.Blur()
			m.textInput.Placeholder = "YYYY-MM-DD"
			m.textInput.SetValue(time.Now().Format("2006-01-02"))
			m.textInput.Focus()
		}
		return m, cmd

	case stageDateInput:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			d, err := time.Parse("2006-01-02", m.textInput.Value())
			if err != nil {
				// ignore parse error; keep focus for correction
				break
			}
			m.session.Date = d
			m.stage = stageDurationInput
			m.textInput.Placeholder = "Duration (minutes)"
			m.textInput.SetValue("")
			m.textInput.Focus()
		}
		return m, cmd

	case stageDurationInput:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			if v, err := strconv.Atoi(m.textInput.Value()); err == nil {
				m.session.Duration = v
				m.stage = stageScoreInput
				m.textInput.Placeholder = "Score (1–5)"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, cmd

	case stageScoreInput:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			if v, err := strconv.Atoi(m.textInput.Value()); err == nil {
				m.session.Score = v
				m.stage = stageNotesInput
				m.textInput.Placeholder = "Notes (optional)"
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}
		return m, cmd

	case stageNotesInput:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			m.session.Notes = m.textInput.Value()
			m.stage = stageDone
			return m, tea.Quit
		}
		return m, cmd

	default:
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI for the current stage
func (m model) View() string {
	switch m.stage {
	case stageSelectTask:
		return m.list.View()
	case stageNewTaskName:
		return fmt.Sprintf(
			"Create a new task:\n\n%s\n\n(enter to confirm)",
			m.textInput.View(),
		)
	case stageDateInput, stageDurationInput, stageScoreInput, stageNotesInput:
		prompt := map[int]string{
			stageDateInput:     "Session date (YYYY-MM-DD):",
			stageDurationInput: "Session duration (minutes):",
			stageScoreInput:    "Session score (1–5):",
			stageNotesInput:    "Session notes (optional):",
		}[m.stage]
		return fmt.Sprintf(
			"%s\n\n%s\n\n(enter to confirm)",
			prompt,
			m.textInput.View(),
		)
	default:
		return ""
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
