package cli

import (
	"database/sql"
	"io"
	"os"

	"github.com/kiki-ki/go-qo/internal/output"
)

// Options configures CLI execution.
type Options struct {
	Format output.Format
	Output io.Writer
}

// DefaultOptions returns default CLI options.
func DefaultOptions() *Options {
	return &Options{
		Format: output.FormatTable,
		Output: os.Stdout,
	}
}

// Run executes a SQL query and prints results.
func Run(db *sql.DB, query string, opts *Options) error {
	if opts == nil {
		opts = DefaultOptions()
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	p := output.NewPrinter(&output.Options{
		Format: opts.Format,
		Output: opts.Output,
	})

	return p.PrintRows(rows)
}
