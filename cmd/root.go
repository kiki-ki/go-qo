// Package cmd provides the CLI commands for qo.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
	"github.com/kiki-ki/go-qo/internal/printer"
)

var (
	outputFormat string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "qo <file1> [file2...] <sql-query>",
	Short: "Execute SQL queries on JSON files",
	Long: `qo is a command-line tool that allows you to query JSON files using SQL.

Examples:
  qo tests.json "SELECT * FROM tests"
  qo companies.json users.json "SELECT c.name, u.name FROM companies c JOIN users u ON c.id = u.company_id"
  qo data.json "SELECT * FROM data WHERE age > 30" --format json`,
	Args: cobra.MinimumNArgs(2),

	RunE: runQuery,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format: table | json | csv")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show informational messages")
}

func runQuery(cmd *cobra.Command, args []string) error {
	query := args[len(args)-1]
	filePaths := args[:len(args)-1]

	// Initialize database
	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	// Load files into database
	if err := loadFiles(database, filePaths); err != nil {
		return err
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

func loadFiles(database *db.DB, filePaths []string) error {
	if verbose {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "### Tables ###")
		fmt.Fprintln(os.Stderr, "")
	}

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

	if verbose {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "### Result ###")
		fmt.Fprintln(os.Stderr, "")
	}

	return nil
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
