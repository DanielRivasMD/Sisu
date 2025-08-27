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

// ////////////////////////////////////////////////////////////////////////////////////////////////////

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"strconv"
// 	"time"

// 	"github.com/aarondl/null/v8"
// 	"github.com/aarondl/sqlboiler/v4/boil"
// 	qm "github.com/aarondl/sqlboiler/v4/queries/qm"
// 	"github.com/charmbracelet/bubbles/list"
// 	"github.com/charmbracelet/bubbles/textinput"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/spf13/cobra"

// 	"github.com/DanielRivasMD/Sisu/db"
// 	"github.com/DanielRivasMD/Sisu/models"
// )

// ////////////////////////////////////////////////////////////////////////////////////////////////////

// // TODO: update for generic implementation carefully
// // TODO: bootstrap `sisu session add = sisu track`
// var trackCmd = &cobra.Command{
// 	Use:     "track",
// 	Short:   "Record a session against a task",
// 	Long:    helpTrack,
// 	Example: exampleTrack,

// 	PreRun:  preRunTrack,
// 	Run:     runTrack,
// 	PostRun: postRunTrack,
// }

// ////////////////////////////////////////////////////////////////////////////////////////////////////

// func init() {
// 	rootCmd.AddCommand(trackCmd)
// }

// ////////////////////////////////////////////////////////////////////////////////////////////////////

// func preRunTrack(cmd *cobra.Command, args []string) {
// 	if _, err := db.InitDB(dbPath); err != nil {
// 		log.Fatalf("database initialization failed: %v", err)
// 	}
// }

// func runTrack(cmd *cobra.Command, args []string) {
// 	// ensure DB is initialized
// 	if db.Conn == nil {
// 		log.Fatalln("database not initialized")
// 	}
// 	ctx := db.Ctx()

// 	// load tasks with SQLBoiler
// 	tasksModels, err := models.Tasks(
// 		qm.Where("archived = ?", false),
// 		qm.OrderBy("name"),
// 	).All(ctx, db.Conn)
// 	if err != nil {
// 		log.Fatalf("fetch tasks: %v", err)
// 	}

// 	// convert to list items
// 	var tasks []taskItem
// 	for _, t := range tasksModels {
// 		tasks = append(tasks, taskItem{
// 			id:   t.ID.Int64,
// 			name: t.Name,
// 		})
// 	}
// 	// add the "create new task" option
// 	tasks = append(tasks, taskItem{id: 0, name: "[Create new task]"})

// 	// start TUI
// 	m := initialModel(tasks)
// 	p := tea.NewProgram(m)
// 	final, err := p.StartReturningModel()
// 	if err != nil {
// 		log.Fatalf("starting TUI: %v", err)
// 	}
// 	m = final.(model)

// 	// if user created a new task, insert it
// 	if m.newTaskName != "" {
// 		newTask := &models.Task{
// 			Name:      m.newTaskName,
// 			DateStart: null.TimeFrom(time.Now()),
// 		}
// 		if err := newTask.Insert(ctx, db.Conn, boil.Infer()); err != nil {
// 			log.Fatalf("insert new task: %v", err)
// 		}
// 		m.selectedID = newTask.ID.Int64
// 	}

// 	// record the session
// 	session := &models.Session{
// 		Task:          m.selectedID,
// 		Date:          m.session.Date,
// 		DurationMins:  null.Int64From(int64(m.session.Duration)),
// 		ScoreFeedback: null.Int64From(int64(m.session.Score)),
// 		Notes:         null.StringFrom(m.session.Notes),
// 	}
// 	if err := session.Insert(ctx, db.Conn, boil.Infer()); err != nil {
// 		log.Fatalf("insert session: %v", err)
// 	}

// 	fmt.Println("\nSession recorded!")
// }

// func postRunTrack(cmd *cobra.Command, args []string) {
// 	if db.Conn != nil {
// 		if err := db.Conn.Close(); err != nil {
// 			fmt.Fprintf(os.Stderr, "error closing database: %v\n", err)
// 		}
// 		db.Conn = nil
// 	}
// }

// ////////////////////////////////////////////////////////////////////////////////////////////////////

// var (
// 	listWidth  = 40
// 	listHeight = 10
// )

// type sessionData struct {
// 	Date     time.Time
// 	Duration int
// 	Score    int
// 	Notes    string
// }

// type taskItem struct {
// 	id   int64
// 	name string
// }

// func (t taskItem) Title() string       { return t.name }
// func (t taskItem) Description() string { return "" }
// func (t taskItem) FilterValue() string { return t.name }

// const (
// 	stageSelectTask = iota
// 	stageNewTaskName
// 	stageDateInput
// 	stageDurationInput
// 	stageScoreInput
// 	stageNotesInput
// 	stageDone
// )

// type model struct {
// 	stage       int
// 	list        list.Model
// 	textInput   textinput.Model
// 	tasks       []taskItem
// 	selectedID  int64
// 	newTaskName string
// 	session     sessionData
// }

// func initialModel(tasks []taskItem) model {
// 	items := make([]list.Item, len(tasks))
// 	for i, t := range tasks {
// 		items[i] = t
// 	}

// 	l := list.New(items, list.NewDefaultDelegate(), listWidth, listHeight)
// 	l.Title = "Select a task:"
// 	l.SetShowHelp(false)

// 	ti := textinput.New()
// 	ti.CharLimit = 64
// 	ti.Width = listWidth

// 	return model{
// 		stage:     stageSelectTask,
// 		list:      l,
// 		textInput: ti,
// 		tasks:     tasks,
// 	}
// }

// func (m model) Init() tea.Cmd { return nil }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch m.stage {
// 	case stageSelectTask:
// 		var cmd tea.Cmd
// 		m.list, cmd = m.list.Update(msg)

// 		switch msg := msg.(type) {
// 		case tea.KeyMsg:
// 			if msg.String() == "enter" {
// 				sel := m.list.SelectedItem().(taskItem)
// 				if sel.id == 0 {
// 					m.stage = stageNewTaskName
// 					m.textInput.Placeholder = "New task name"
// 					m.textInput.SetValue("")
// 					m.textInput.Focus()
// 				} else {
// 					m.selectedID = sel.id
// 					m.stage = stageDateInput
// 					m.textInput.Placeholder = "YYYY-MM-DD"
// 					m.textInput.SetValue(time.Now().Format("2006-01-02"))
// 					m.textInput.Focus()
// 				}
// 			}
// 			return m, nil
// 		}
// 		return m, cmd

// 	case stageNewTaskName:
// 		ti, cmd := m.textInput.Update(msg)
// 		m.textInput = ti
// 		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
// 			m.newTaskName = m.textInput.Value()
// 			m.stage = stageDateInput
// 			m.textInput.Placeholder = "YYYY-MM-DD"
// 			m.textInput.SetValue(time.Now().Format("2006-01-02"))
// 			m.textInput.Focus()
// 		}
// 		return m, cmd

// 	case stageDateInput:
// 		ti, cmd := m.textInput.Update(msg)
// 		m.textInput = ti
// 		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
// 			d, err := time.Parse("2006-01-02", m.textInput.Value())
// 			if err == nil {
// 				m.session.Date = d
// 				m.stage = stageDurationInput
// 				m.textInput.Placeholder = "Duration (minutes)"
// 				m.textInput.SetValue("")
// 				m.textInput.Focus()
// 			}
// 		}
// 		return m, cmd

// 	case stageDurationInput:
// 		ti, cmd := m.textInput.Update(msg)
// 		m.textInput = ti
// 		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
// 			if v, err := strconv.Atoi(m.textInput.Value()); err == nil {
// 				m.session.Duration = v
// 				m.stage = stageScoreInput
// 				m.textInput.Placeholder = "Score (1–5)"
// 				m.textInput.SetValue("")
// 				m.textInput.Focus()
// 			}
// 		}
// 		return m, cmd

// 	case stageScoreInput:
// 		ti, cmd := m.textInput.Update(msg)
// 		m.textInput = ti
// 		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
// 			if v, err := strconv.Atoi(m.textInput.Value()); err == nil {
// 				m.session.Score = v
// 				m.stage = stageNotesInput
// 				m.textInput.Placeholder = "Notes (optional)"
// 				m.textInput.SetValue("")
// 				m.textInput.Focus()
// 			}
// 		}
// 		return m, cmd

// 	case stageNotesInput:
// 		ti, cmd := m.textInput.Update(msg)
// 		m.textInput = ti
// 		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
// 			m.session.Notes = m.textInput.Value()
// 			m.stage = stageDone
// 			return m, tea.Quit
// 		}
// 		return m, cmd

// 	default:
// 		return m, tea.Quit
// 	}
// }

// func (m model) View() string {
// 	switch m.stage {
// 	case stageSelectTask:
// 		return m.list.View()
// 	case stageNewTaskName:
// 		return fmt.Sprintf(
// 			"Create a new task:\n\n%s\n\n(enter to confirm)",
// 			m.textInput.View(),
// 		)
// 	case stageDateInput, stageDurationInput, stageScoreInput, stageNotesInput:
// 		prompt := map[int]string{
// 			stageDateInput:     "Session date (YYYY-MM-DD):",
// 			stageDurationInput: "Session duration (minutes):",
// 			stageScoreInput:    "Session score (1–5):",
// 			stageNotesInput:    "Session notes (optional):",
// 		}[m.stage]
// 		return fmt.Sprintf(
// 			"%s\n\n%s\n\n(enter to confirm)",
// 			prompt,
// 			m.textInput.View(),
// 		)
// 	default:
// 		return ""
// 	}
// }

// ////////////////////////////////////////////////////////////////////////////////////////////////////
