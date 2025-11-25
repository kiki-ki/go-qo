// Package cmd provides the CLI commands for qo.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
	"github.com/kiki-ki/go-qo/internal/printer"
)

const stdinTableName = "t"

var (
	outputFormat string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "qo [file1 file2...] <sql-query>",
	Short: "Execute SQL queries on JSON files",
	Long: `qo is a command-line tool that allows you to query JSON files using SQL.

Examples:
  qo tests.json "SELECT * FROM tests"
  qo companies.json users.json "SELECT c.name, u.name FROM companies c JOIN users u ON c.id = u.company_id"
  qo data.json "SELECT * FROM data WHERE age > 30" --format json

Pipe from stdin (table name is 't'):
  curl https://api.example.com/users | qo "SELECT * FROM t"
  cat data.json | qo "SELECT name, age FROM t WHERE age > 30"`,
	Args: cobra.MinimumNArgs(1),

	RunE: runQuery,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format: table | json | csv")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show informational messages")
}

func runQuery(cmd *cobra.Command, args []string) error {
	// Initialize database
	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	// Check if stdin has data
	stdinHasData, err := hasStdinData()
	if err != nil {
		return err
	}

	var query string
	var filePaths []string

	if stdinHasData {
		// stdin mode: only query is required
		query = args[len(args)-1]
		filePaths = args[:len(args)-1]

		// Load stdin data
		if err := loadStdin(database); err != nil {
			return err
		}
	} else {
		// file mode: at least one file and query required
		if len(args) < 2 {
			return fmt.Errorf("requires at least one file and a query, or pipe data via stdin")
		}
		query = args[len(args)-1]
		filePaths = args[:len(args)-1]
	}

	// Load files into database (if any)
	if len(filePaths) > 0 {
		if err := loadFiles(database, filePaths); err != nil {
			return err
		}
	}

	// Execute query
	rows, err := database.Query(query)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Print results
	p := printer.New(&printer.Options{
		Format: printer.Format(outputFormat),
		Output: os.Stdout,
	})

	if err := p.PrintRows(rows); err != nil {
		return fmt.Errorf("failed to print results: %w", err)
	}

	return nil
}

// hasStdinData checks if there's data available on stdin.
func hasStdinData() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to stat stdin: %w", err)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0, nil
}

// loadStdin loads JSON data from stdin into the database.
func loadStdin(database *db.DB) error {
	if verbose {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "### Tables ###")
		fmt.Fprintln(os.Stderr, "")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	parsed, err := parser.ParseJSONBytes(data)
	if err != nil {
		return fmt.Errorf("failed to parse stdin: %w", err)
	}

	if err := database.LoadData(stdinTableName, parsed); err != nil {
		return fmt.Errorf("failed to load stdin data: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "- stdin → %s (%d rows, %d columns)\n",
			stdinTableName, len(parsed.Rows), len(parsed.Columns))
	}

	return nil
}

func loadFiles(database *db.DB, filePaths []string) error {
	for _, path := range filePaths {
		data, err := parser.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		tableName := db.TableNameFromPath(path)

		if err := database.LoadData(tableName, data); err != nil {
			return fmt.Errorf("failed to load table %s: %w", tableName, err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "- %s → %s (%d rows, %d columns)\n",
				path, tableName, len(data.Rows), len(data.Columns))
		}
	}

	return nil
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
