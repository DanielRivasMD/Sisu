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

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Manage review entries (list, rm via CLI; add, edit via TUI)",
}

var reviewAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive TUI to add a new review",
	Run:   runReviewAdd,
}

var reviewEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Interactive TUI to edit an existing review",
	Args:  cobra.ExactArgs(1),
	Run:   runReviewEdit,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	// attach the review parent and TUI subcommands
	rootCmd.AddCommand(reviewCmd)
	reviewCmd.AddCommand(reviewAddCmd, reviewEditCmd)

	RegisterCrudSubcommands(reviewCmd, "sisu.db", CrudModel[*models.Review]{
		Singular: "review",

		ListFn: func(ctx context.Context, conn *sql.DB) ([]*models.Review, error) {
			return models.Reviews(qm.OrderBy("id ASC")).All(ctx, conn)
		},

		Format: func(r *models.Review) (int64, string) {
			wk := ""
			if r.Week.Valid {
				wk = strconv.FormatInt(r.Week.Int64, 10)
			}
			return r.ID.Int64, fmt.Sprintf("task=%d week=%s summary=%s", r.Task, wk, r.Summary.String)
		},

		TableHeaders: []string{"id", "task", "week", "summary"},
		TableRow: func(r *models.Review) []string {
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

		RemoveFn: func(ctx context.Context, conn *sql.DB, id int64) error {
			row, err := models.FindReview(ctx, conn, null.Int64From(id))
			if err != nil {
				return err
			}
			_, err = row.Delete(ctx, conn)
			return err
		},
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runReviewAdd(_ *cobra.Command, _ []string) {
	rev := &models.Review{}

	fields := []Field{
		FInt("Task ID", "Task", ""),
		FOptInt("Week (optional)", "Week", ""),
		FOptString("Summary (optional)", "Summary", ""),
	}

	RunFormWizard(fields, rev)

	if err := rev.Insert(context.Background(), db.Conn, boil.Infer()); err != nil {
		log.Fatalf("insert review: %v", err)
	}
	fmt.Printf("Created review %d\n", rev.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runReviewEdit(_ *cobra.Command, args []string) {
	rawID := args[0]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid review ID %q: %v", rawID, err)
	}

	rev, err := models.FindReview(context.Background(), db.Conn, null.Int64From(idNum))
	if err != nil {
		log.Fatalf("find review %d: %v", idNum, err)
	}

	fields := []Field{
		FInt("Task ID", "Task", strconv.FormatInt(rev.Task, 10)),
		FOptInt("Week (optional)", "Week", OptInt64Initial(rev.Week)),
		FOptString("Summary (optional)", "Summary", rev.Summary.String),
	}

	RunFormWizard(fields, rev)

	if _, err := rev.Update(context.Background(), db.Conn, boil.Whitelist("task", "week", "summary")); err != nil {
		log.Fatalf("update review: %v", err)
	}
	fmt.Printf("Updated review %d\n", rev.ID.Int64)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
