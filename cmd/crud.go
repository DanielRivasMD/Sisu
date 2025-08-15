// cmd/crud.go
package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"database/sql"
	"github.com/spf13/cobra"

	"github.com/DanielRivasMD/Sisu/db"
)

// CrudModel describes everything that varies per table.
type CrudModel[T any] struct {
	Singular string
	ListFn   func(ctx context.Context, db *sql.DB) ([]T, error)
	Format   func(item T) (int64, string)
	AddFn    func(ctx context.Context, db *sql.DB, args []string) (int64, error)
	RemoveFn func(ctx context.Context, db *sql.DB, id int64) error
	EditFn   func(ctx context.Context, db *sql.DB, id int64, args []string) error
}

// NewCrudCmd builds `root <singular>` with `list`, `add`, `rm`, `edit`.
func NewCrudCmd[T any](
	dbPath string,
	desc CrudModel[T],
) *cobra.Command {
	parent := &cobra.Command{
		Use:   desc.Singular,
		Short: fmt.Sprintf("CRUD for %s", desc.Singular),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if _, err := db.InitDB(dbPath); err != nil {
				log.Fatalf("init DB: %v", err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if db.Conn != nil {
				_ = db.Conn.Close()
				db.Conn = nil
			}
		},
	}

	// list
	parent.AddCommand(&cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List all %s", desc.Singular),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := db.Ctx()
			items, err := desc.ListFn(ctx, db.Conn)
			if err != nil {
				log.Fatalf("list %s: %v", desc.Singular, err)
			}
			for _, it := range items {
				id, human := desc.Format(it)
				fmt.Printf("%d\t%s\n", id, human)
			}
		},
	})

	// add
	parent.AddCommand(&cobra.Command{
		Use:   "add [fields...]",
		Short: fmt.Sprintf("Add a new %s", desc.Singular),
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := db.Ctx()
			id, err := desc.AddFn(ctx, db.Conn, args)
			if err != nil {
				log.Fatalf("add %s: %v", desc.Singular, err)
			}
			fmt.Printf("✅ Created %s %d\n", desc.Singular, id)
		},
	})

	// rm
	rm := &cobra.Command{
		Use:   "rm [id]",
		Short: fmt.Sprintf("Remove a %s by ID", desc.Singular),
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			raw, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatalf("invalid id: %v", err)
			}
			ctx := db.Ctx()
			if err := desc.RemoveFn(ctx, db.Conn, raw); err != nil {
				log.Fatalf("rm %s: %v", desc.Singular, err)
			}
			fmt.Printf("🗑️  Removed %s %d\n", desc.Singular, raw)
		},
		// optional: live completion of IDs
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := db.Ctx()
			items, err := desc.ListFn(ctx, db.Conn)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			var comps []string
			for _, it := range items {
				id, _ := desc.Format(it)
				s := strconv.FormatInt(id, 10)
				if toComplete == "" || strings.HasPrefix(s, toComplete) {
					comps = append(comps, s)
				}
			}
			return comps, cobra.ShellCompDirectiveNoFileComp
		},
	}
	parent.AddCommand(rm)

	// edit (if provided)
	if desc.EditFn != nil {
		parent.AddCommand(&cobra.Command{
			Use:   "edit [id] [fields...]",
			Short: fmt.Sprintf("Edit a %s by ID", desc.Singular),
			Args:  cobra.MinimumNArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				raw, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					log.Fatalf("invalid id: %v", err)
				}
				ctx := db.Ctx()
				if err := desc.EditFn(ctx, db.Conn, raw, args[1:]); err != nil {
					log.Fatalf("edit %s: %v", desc.Singular, err)
				}
				fmt.Printf("✏️  Updated %s %d\n", desc.Singular, raw)
			},
		})
	}

	return parent
}
