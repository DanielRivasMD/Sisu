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
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

type CrudModel[T any] struct {
	Singular string

	// Required
	ListFn   func(ctx context.Context, db *sql.DB) ([]T, error)
	RemoveFn func(ctx context.Context, db *sql.DB, id int64) error

	// Legacy single-line formatting fallback
	Format func(item T) (int64, string)

	// Optional: rich table output
	// If both are set, `list` prints a table using these.
	TableHeaders []string
	TableRow     func(item T) []string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func RegisterCrudSubcommands[T any](
	parent *cobra.Command,
	dbPath string,
	desc CrudModel[T],
) {
	parent.PersistentPreRun = persistentPreRun
	parent.PersistentPostRun = persistentPostRun

	// list
	list := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List all %s", desc.Singular),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := db.Ctx()
			items, err := desc.ListFn(ctx, db.Conn)
			if err != nil {
				log.Fatalf("list %s: %v", desc.Singular, err)
			}

			if desc.TableHeaders != nil && desc.TableRow != nil {
				// Render as a table
				rows := make([][]string, 0, len(items))
				for _, it := range items {
					rows = append(rows, desc.TableRow(it))
				}
				fmt.Println(RenderTable(desc.TableHeaders, rows))
				return
			}

			// Fallback: legacy one-line format
			for _, it := range items {
				id, human := desc.Format(it)
				fmt.Printf("%d\t%s\n", id, human)
			}
		},
	}
	parent.AddCommand(list)

	// rm
	rm := &cobra.Command{
		Use:   "rm [id]...",
		Short: fmt.Sprintf("Remove a %s by ID", desc.Singular),
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, a := range args {
				raw, err := strconv.ParseInt(a, 10, 64)
				if err != nil {
					log.Fatalf("invalid id: %v", err)
				}
				ctx := db.Ctx()
				if err := desc.RemoveFn(ctx, db.Conn, raw); err != nil {
					log.Fatalf("rm %s: %v", desc.Singular, err)
				}
				fmt.Printf("Removed %s %d\n", desc.Singular, raw)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Only provide completion if we have a formatter to show human hints
			if desc.Format == nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			// Exclude already provided IDs (rm supports multiple)
			used := make(map[string]struct{}, len(args))
			for _, a := range args {
				used[a] = struct{}{}
			}
			ctx := db.Ctx()
			return buildIDCompletions(ctx, db.Conn, desc.ListFn, desc.Format, toComplete, used)
		},
	}
	parent.AddCommand(rm)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// buildIDCompletions returns completions like "123\tname=foo start=..."
// - toComplete: user’s current partial input
// - used: set of IDs already typed (for multi-arg rm)
func buildIDCompletions[T any](
	ctx context.Context,
	dbConn *sql.DB,
	listFn func(ctx context.Context, db *sql.DB) ([]T, error),
	format func(item T) (int64, string),
	toComplete string,
	used map[string]struct{},
) ([]string, cobra.ShellCompDirective) {
	// Ensure DB for completion (PersistentPreRun doesn’t run for __complete)
	if err := EnsureDB(); err != nil {
		// Don’t panic; tell the shell there are no completions and no file fallback.
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Use the (now ensured) connection from your package, not the stale dbConn arg
	items, err := listFn(ctx, db.Conn)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	toComplete = strings.TrimSpace(toComplete)
	comps := make([]string, 0, len(items))
	for _, it := range items {
		id, human := format(it)
		s := strconv.FormatInt(id, 10)
		if used != nil {
			if _, skip := used[s]; skip {
				continue
			}
		}
		if toComplete != "" && !strings.HasPrefix(s, toComplete) {
			continue
		}
		comps = append(comps, s+"\t"+human)
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func AttachEditCompletion[T any](
	cmd *cobra.Command,
	listFn func(ctx context.Context, db *sql.DB) ([]T, error),
	format func(item T) (int64, string),
) {
	cmd.ValidArgsFunction = func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first positional argument
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if format == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		ctx := db.Ctx()
		return buildIDCompletions(ctx, db.Conn, listFn, format, toComplete, nil)
	}
}

func AttachRmCompletion[T any](
	cmd *cobra.Command,
	listFn func(ctx context.Context, db *sql.DB) ([]T, error),
	format func(item T) (int64, string),
) {
	cmd.ValidArgsFunction = func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if format == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// Exclude already-provided IDs
		used := make(map[string]struct{}, len(args))
		for _, a := range args {
			used[a] = struct{}{}
		}
		ctx := db.Ctx()
		return buildIDCompletions(ctx, db.Conn, listFn, format, toComplete, used)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// RenderTable creates an ASCII table similar to the example you shared.
func RenderTable(headers []string, rows [][]string) string {
	// Compute column widths
	w := make([]int, len(headers))
	for i, h := range headers {
		w[i] = len(h)
	}
	for _, r := range rows {
		for i := range headers {
			if i < len(r) && len(r[i]) > w[i] {
				w[i] = len(r[i])
			}
		}
	}

	// Builders
	var b strings.Builder
	divider := func() {
		b.WriteString("+")
		for i := range headers {
			b.WriteString(strings.Repeat("-", w[i]+2))
			b.WriteString("+")
		}
		b.WriteString("\n")
	}
	writeRow := func(cols []string) {
		b.WriteString("|")
		for i := range headers {
			val := ""
			if i < len(cols) {
				val = cols[i]
			}
			// pad right
			b.WriteString(" ")
			b.WriteString(val)
			b.WriteString(strings.Repeat(" ", w[i]-len(val)))
			b.WriteString(" ")
			b.WriteString("|")
		}
		b.WriteString("\n")
	}

	divider()
	writeRow(headers)
	divider()
	for _, r := range rows {
		writeRow(r)
	}
	divider()

	// Prefix each line with a pipe-like margin if you want to match your example numbering style
	// Caller can print as-is. We return the clean table string here.
	return b.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type Field struct {
	Name     string
	Label    string
	Initial  string
	Validate func(input string) error
	err      error
	Parse    func(string) (any, error)
	Assign   func(holder any, v any)
	Input    textinput.Model
}

type FormModel struct {
	fields []Field
	idx    int
	holder any
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func NewFormModel(fields []Field, holder any) FormModel {
	for i := range fields {
		ti := textinput.New()
		ti.Placeholder = fields[i].Label
		ti.SetValue(fields[i].Initial)
		if i == 0 {
			ti.Focus()
		}
		fields[i].Input = ti
	}
	return FormModel{
		fields: fields,
		idx:    0,
		holder: holder,
	}
}

func (m FormModel) Init() tea.Cmd { return nil }

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f := &m.fields[m.idx]

	ti, cmd := f.Input.Update(msg)
	f.Input = ti

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		raw := f.Input.Value()

		if f.Validate != nil {
			if err := f.Validate(raw); err != nil {
				f.err = err
				return m, nil
			}
		}

		v, err := f.Parse(raw)
		if err != nil {
			f.err = err
			return m, nil
		}

		f.Assign(m.holder, v)

		f.err = nil
		m.idx++
		if m.idx >= len(m.fields) {
			return m, tea.Quit
		}
		m.fields[m.idx].Input.Focus()
		return m, nil
	}

	return m, cmd
}

func (m FormModel) View() string {
	if m.idx >= len(m.fields) {
		return ""
	}
	f := m.fields[m.idx]
	header := fmt.Sprintf("[%d/%d] %s\n\n", m.idx+1, len(m.fields), f.Label)
	body := f.Input.View()
	errLine := ""
	if f.err != nil {
		errLine = "\n\n! " + f.err.Error()
	}
	footer := "\n\n(enter to confirm, ctrl+c to cancel)"
	return header + body + errLine + footer
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func RunFormWizard(fields []Field, holder any) {
	p := tea.NewProgram(NewFormModel(fields, holder))
	if _, err := p.StartReturningModel(); err != nil {
		log.Fatalf("form wizard failed: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func RunFormWizardWithSubmit(fields []Field, holder any, onSubmit func(holder any) error) {
	p := tea.NewProgram(NewFormModel(fields, holder))
	if _, err := p.StartReturningModel(); err != nil {
		log.Fatalf("form wizard failed: %v", err)
	}
	if err := onSubmit(holder); err != nil {
		log.Fatalf("submit failed: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
