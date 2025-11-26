// Package cmd provides the CLI commands for qo.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kiki-ki/go-qo/internal/cli"
	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/input"
	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/internal/tui"
)

const stdinTableName = "tmp"

var (
	outputFormat string
	inputFormat  string
	queryFlag    string
)

var rootCmd = &cobra.Command{
	Use:   "qo [files...]",
	Short: "Execute SQL queries on JSON files",
	Long: `qo is a command-line tool that allows you to query JSON files using SQL.

TUI Mode (default):
	qo data.json
	cat data.json | qo

CLI Mode (with -q flag):
	qo -q "SELECT * FROM data" data.json
	cat data.json | qo -q "SELECT * FROM t"`,

	Args: cobra.ArbitraryArgs,
	RunE: runQuery,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table (default) | json | csv")
	rootCmd.Flags().StringVarP(&inputFormat, "input", "i", "json", "Input format: json (default)")
	rootCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "SQL query to execute (if omitted, TUI mode)")
}

// runConfig holds the parsed configuration for a query run.
type runConfig struct {
	query      string
	filePaths  []string
	tableNames []string
}

func runQuery(cmd *cobra.Command, args []string) error {
	if err := run(cmd, args); err != nil {
		return fmt.Errorf("%w\n", err)
	}
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	if err := validateFormats(); err != nil {
		return err
	}

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	loader := input.NewLoader(database, input.Format(inputFormat))

	hasStdinData, err := input.HasStdinData()
	if err != nil {
		return err
	}

	cfg := &runConfig{
		query:     queryFlag,
		filePaths: args,
	}

	if err := loadData(loader, cfg, hasStdinData); err != nil {
		return err
	}

	return execute(database, cfg)
}

// validateFormats checks if input/output formats are valid.
func validateFormats() error {
	if !input.IsValidFormat(inputFormat) {
		return fmt.Errorf("unsupported input format: %s (supported: %v)", inputFormat, input.Formats())
	}
	if !output.IsValidFormat(outputFormat) {
		return fmt.Errorf("unsupported output format: %s (supported: %v)", outputFormat, output.Formats())
	}
	return nil
}

// loadData loads data from stdin and/or files into the database.
// Returns the list of loaded table names.
func loadData(loader *input.Loader, cfg *runConfig, hasStdinData bool) error {
	if !hasStdinData && len(cfg.filePaths) == 0 {
		return fmt.Errorf("no input data: provide files as arguments or pipe data via stdin")
	}

	var tableNames []string

	if hasStdinData {
		if err := loader.LoadStdin(stdinTableName); err != nil {
			return err
		}
		tableNames = append(tableNames, stdinTableName)
	}

	if len(cfg.filePaths) > 0 {
		if err := loader.LoadFiles(cfg.filePaths); err != nil {
			return err
		}
		for _, path := range cfg.filePaths {
			tableNames = append(tableNames, db.TableNameFromPath(path))
		}
	}

	cfg.tableNames = tableNames
	return nil
}

// execute runs either TUI or CLI mode based on configuration.
// TUI mode is used when query is empty, CLI mode when query is provided via -q flag.
func execute(database *db.DB, cfg *runConfig) error {
	if cfg.query == "" {
		result, err := tui.Run(database.DB, cfg.tableNames)
		if err != nil {
			return err
		}
		if result == nil || result.Query == "" {
			return nil
		}
		cfg.query = result.Query
	}
	return cli.Run(database.DB, cfg.query, &cli.Options{
		Format: output.Format(outputFormat),
		Output: os.Stdout,
	})
}

func Execute() error {
	return rootCmd.Execute()
}
