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

var coachCmd = &cobra.Command{
	Use:   "coach",
	Short: "Manage coach triggers (list, rm via CLI; add/edit via TUI)",
}

var coachAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new coach entry",
	Run:   runCoachAdd,
}

var coachEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit a coach entry",
	Args:  cobra.ExactArgs(1),
	Run:   runCoachEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(coachCmd)
	coachCmd.AddCommand(coachAddCmd, coachEditCmd)

	RegisterCrudSubcommands(coachCmd, "sisu.db", CrudModel[*models.Coach]{
		Singular: "coach",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Coach, error) {
			return models.Coaches(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(c *models.Coach) (int64, string) {
			date := ""
			date = c.Date.Format("2006-01-02")
			return c.ID.Int64,
				fmt.Sprintf("trigger=%s content=%s date=%s",
					c.Trigger, c.Content, date,
				)
		},

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			c, err := models.FindCoach(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = c.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runCoachAdd launches a TUI to create a new coach record
func runCoachAdd(_ *cobra.Command, _ []string) {
	entry := &models.Coach{}

	fields := []Field{
		{
			Label:   "Trigger (non-empty)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return nil, fmt.Errorf("trigger cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Trigger").
					SetString(v.(string))
			},
		},
		{
			Label:   "Content (non-empty)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return nil, fmt.Errorf("content cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Content").
					SetString(v.(string))
			},
		},
		{
			Label:   "Date (YYYY-MM-DD, optional)",
			Initial: "",
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return null.Time{}, nil
				}
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
	}

	RunFormWizard(fields, entry)

	if err := entry.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert coach entry: %v", err)
	}
	fmt.Printf("✅ Created coach %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// runCoachEdit launches a pre-seeded TUI then UPDATEs the record
func runCoachEdit(_ *cobra.Command, args []string) {
	rawID := args[0]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid coach ID %q: %v", rawID, err)
	}

	entry, err := models.FindCoach(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find coach %d: %v", idNum, err)
	}

	fields := []Field{
		{
			Label:   "Trigger",
			Initial: entry.Trigger,
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return nil, fmt.Errorf("trigger cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Trigger").
					SetString(v.(string))
			},
		},
		{
			Label:   "Content",
			Initial: entry.Content,
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return nil, fmt.Errorf("content cannot be blank")
				}
				return s, nil
			},
			Assign: func(holder interface{}, v interface{}) {
				reflect.ValueOf(holder).
					Elem().
					FieldByName("Content").
					SetString(v.(string))
			},
		},
		{
			Label: "Date (YYYY-MM-DD, optional)",
			Initial: func() string {
				return entry.Date.Format("2006-01-02")
			}(),
			Parse: func(s string) (interface{}, error) {
				if s == "" {
					return null.Time{}, nil
				}
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
	}

	RunFormWizard(fields, entry)

	if _, err := entry.Update(context.Background(), db.Conn, boil.Whitelist("trigger", "content", "date")); err != nil {
		log.Fatalf("update coach entry: %v", err)
	}
	fmt.Printf("✅ Updated coach %d\n", entry.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
