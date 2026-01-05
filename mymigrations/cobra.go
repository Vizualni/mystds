package mymigrations

import (
	"context"
	"io/fs"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

const MaxMigrationVersion = int64((1 << 63) - 1)

func runMigrations(ctx context.Context, db *sqlx.DB, version int64) error {
	// the default value is the largest possible int64 value.
	// that is the copy-pasted default value from the goose library.
	if version == 0 {
		version = MaxMigrationVersion
	}
	if err := goose.UpToContext(ctx, db.DB, ".", version); err != nil {
		return err
	}
	return nil
}

func NewCobraCommand(
	name string,
	dialect string,
	fs fs.FS,
	db func() *sqlx.DB,
) *cobra.Command {
	goose.SetBaseFS(fs)
	goose.SetDialect(dialect)

	cmd := &cobra.Command{
		Use:   name + "migrations",
		Short: "Run migrations for " + name,
	}

	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Run migrations up",
		Long:  "Run migrations up to the latest version or up to the version specified in the first argument",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			upTo := MaxMigrationVersion
			if len(args) > 0 {
				var err error
				upTo, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return err
				}
			}
			return runMigrations(cmd.Context(), db(), upTo)
		},
	}

	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Run migrations down",
		Long:  "Run migrations down to the latest version or down the number of migrations specified in the first argument",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// default to 1
			num := 1
			var err error
			if len(args) > 0 {
				num, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
			}
			for i := 0; i < num; i++ {
				if err := goose.Down(db().DB, "."); err != nil {
					return err
				}
			}
			return nil
		},
	}

	newMigrationCmd := &cobra.Command{
		Use:     "new",
		Aliases: []string{"create"},
		Short:   "Create new migration",
		Long:    "Create new migration file with the name specified in the first argument",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return goose.Create(db().DB, "./", args[0], "sql")
		},
	}

	statusMigrationCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return goose.Status(db().DB, ".")
		},
	}

	cmd.AddCommand(upCmd, downCmd, newMigrationCmd, statusMigrationCmd)
	return cmd
}
