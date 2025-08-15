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
	"strconv"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
	"github.com/DanielRivasMD/Sisu/models"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// taskCmd is the parent for subcommands: list, add (TUI), rm.
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks (list, add, rm)",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := db.InitDB(dbPath); err != nil {
			log.Fatalf("failed to init DB: %v", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db.Conn != nil {
			_ = db.Conn.Close()
			db.Conn = nil
		}
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all non-archived tasks",
	Run:   runTaskList,
}

var taskRmCmd = &cobra.Command{
	Use:               "rm [id]",
	Short:             "Remove a task by ID",
	Args:              cobra.ExactArgs(1),
	Run:               runTaskRm,
	ValidArgsFunction: taskRmCompletions,
}

var taskAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new task",
	Run:   runTaskAddTUI,
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd, taskAddCmd, taskRmCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runTaskList prints ID, name, description, target date, and archived flag.
func runTaskList(cmd *cobra.Command, args []string) {
	ctx := db.Ctx()
	tasks, err := models.Tasks(
		qm.OrderBy("id ASC"),
	).All(ctx, db.Conn)
	if err != nil {
		log.Fatalf("query tasks: %v", err)
	}

	if len(tasks) == 0 {
		fmt.Println("no tasks found")
		return
	}

	fmt.Printf("ID\tName\tDescription\tTarget\tArchived\n")
	fmt.Printf("--\t----\t-----------\t------\t--------\n")
	for _, t := range tasks {
		desc := t.Description.String
		target := ""
		if !t.DateTarget.Time.IsZero() {
			target = t.DateTarget.Time.Format("2006-01-02")
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%v\n",
			t.ID.Int64,
			t.Name,
			desc,
			target,
			t.Archived.Bool,
		)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runTaskRm deletes the task with the given ID.
func runTaskRm(cmd *cobra.Command, args []string) {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Fatalf("invalid task id: %v", err)
	}
	ctx := db.Ctx()

	task, err := models.FindTask(ctx, db.Conn, null.Int64From(id))
	if err != nil {
		log.Fatalf("find task: %v", err)
	}

	if _, err := task.Delete(ctx, db.Conn); err != nil {
		log.Fatalf("delete task: %v", err)
	}

	fmt.Printf("Removed task %d: %s\n", id, task.Name)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// taskRmCompletions provides shell completion of existing task IDs.
func taskRmCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := db.Ctx()
	tasks, err := models.Tasks(qm.OrderBy("id ASC")).All(ctx, db.Conn)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var comps []string
	for _, t := range tasks {
		s := strconv.FormatInt(t.ID.Int64, 10)
		if toComplete == "" || strings.HasPrefix(s, toComplete) {
			comps = append(comps, s)
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	addStageName = iota
	addStageDesc
	addStageTarget
	addStageStart
	addStageDone
)

type addModel struct {
	stage       int
	textInput   textinput.Model
	name        string
	description string
	target      string
}

// runTaskAddTUI launches a Bubble Tea TUI to gather name, description, and target date.
func runTaskAddTUI(cmd *cobra.Command, args []string) {
	p := tea.NewProgram(initialAddModel())
	m, err := p.StartReturningModel()
	if err != nil {
		log.Fatalf("starting add TUI: %v", err)
	}
	m2 := m.(addModel)

	// parse target date, empty means nil
	var dt null.Time
	if t := strings.TrimSpace(m2.target); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			dt = null.TimeFrom(parsed)
		}
	}

	newTask := &models.Task{
		Name:        m2.name,
		Description: null.StringFrom(m2.description),
		DateTarget:  dt,
		// date_start defaults in DB
		Archived: null.BoolFrom(false),
	}

	ctx := db.Ctx()
	if err := newTask.Insert(ctx, db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert task: %v", err)
	}
	fmt.Printf("✅ Created task %d: %s\n", newTask.ID.Int64, newTask.Name)
}

func initialAddModel() addModel {
	ti := textinput.New()
	ti.Placeholder = "Task name"
	ti.Width = 40
	ti.Focus()

	return addModel{
		stage:     addStageName,
		textInput: ti,
	}
}

func (m addModel) Init() tea.Cmd {
	return nil
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ti, cmd := m.textInput.Update(msg)
	m.textInput = ti

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		switch m.stage {
		case addStageName:
			m.name = m.textInput.Value()
			m.stage = addStageDesc
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Description (optional)"
			return m, m.textInput.Focus()
		case addStageDesc:
			m.description = m.textInput.Value()
			m.stage = addStageTarget
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Target date (YYYY-MM-DD)"
			return m, m.textInput.Focus()
		case addStageTarget:
			m.target = m.textInput.Value()
			m.stage = addStageStart
			m.textInput.Placeholder = "YYYY-MM-DD"
			m.textInput.SetValue(time.Now().Format("2006-01-02"))
			return m, m.textInput.Focus()
		case addStageStart:
			m.target = m.textInput.Value()
			m.stage = addStageDone
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m addModel) View() string {
	title := map[int]string{
		addStageName:   "Enter task name:",
		addStageDesc:   "Enter description (optional):",
		addStageTarget: "Enter target date (YYYY-MM-DD):",
		addStageStart:  "Enter start date (YYYY-MM-DD):",
	}[m.stage]
	return fmt.Sprintf(
		"%s\n\n%s\n\n(enter to confirm)\n",
		title,
		m.textInput.View(),
	)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
