// Package cmd provides the CLI commands for qo.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kiki-ki/go-qo/internal/cli"
	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/input"
	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/internal/ui"
)

var version = "dev"

var (
	outputFormat string
	inputFormat  string
	queryFlag    string
	noHeader     bool
)

const stdinTableName = "tmp"

var rootCmd = &cobra.Command{
	Version: version,
	Use:     "qo [files...]",
	Short:   "Query structured data with SQL",
	Long:    "qo is a TUI/CLI tool that lets you query structured data using SQL.",
	Example: strings.Join([]string{
		"  qo data.json                                       # Interactive TUI mode",
		"  cat data.json | qo                                 # Pipe to TUI, output to stdout",
		"  cat data.json | qo - other.json                    # Read stdin alongside files",
		`  qo -q "SELECT * FROM data" data.json               # Direct query mode`,
		`  qo -o json data.csv -q "SELECT * FROM data"        # CSV to JSON`,
		`  qo -q "SELECT * FROM a JOIN b" a.csv b.json        # Join across formats`,
	}, "\n"),
	Args: cobra.ArbitraryArgs,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&inputFormat, "input", "i", "", "Input format: json, csv, tsv, psv (default: by file extension, json if unknown)")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json, jsonl, csv, tsv, psv, table")
	rootCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "SQL query to execute (if omitted, interactive mode)")
	rootCmd.Flags().BoolVar(&noHeader, "no-header", false, "Treat first row as data, not header (CSV/TSV/PSV only)")
}

// runConfig holds the parsed configuration for a query run.
type runConfig struct {
	query      string
	filePaths  []string
	tableNames []string
}

func run(cmd *cobra.Command, args []string) error {
	// Silenced here rather than on the command so that flag parsing errors,
	// which happen before RunE, still show usage. Everything past this point
	// is a runtime failure that usage cannot help with.
	cmd.SilenceUsage = true

	if err := validateFormats(); err != nil {
		return err
	}

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = database.Close() }()

	// An unset -i leaves the format to each input, so files of different
	// formats can be joined.
	loader := input.NewLoader(database, input.Format(inputFormat), &input.LoaderOptions{
		NoHeader: noHeader,
	})

	filePaths, stdinRequested := input.SplitArgs(args)
	useStdin, err := input.UseStdin(filePaths, stdinRequested)
	if err != nil {
		return err
	}

	cfg := &runConfig{
		query:     queryFlag,
		filePaths: filePaths,
	}

	if err := loadData(loader, cfg, useStdin); err != nil {
		return err
	}

	return execute(database, cfg)
}

// validateFormats checks if input/output formats are valid.
func validateFormats() error {
	if inputFormat != "" && !input.IsValidFormat(inputFormat) {
		return fmt.Errorf("unsupported input format: %s (supported: %v)", inputFormat, input.Formats())
	}
	if !output.IsValidFormat(outputFormat) {
		return fmt.Errorf("unsupported output format: %s (supported: %v)", outputFormat, output.Formats())
	}
	return nil
}

// loadData loads data from stdin and/or files into the database.
// Returns the list of loaded table names.
func loadData(loader *input.Loader, cfg *runConfig, useStdin bool) error {
	if !useStdin && len(cfg.filePaths) == 0 {
		return fmt.Errorf("no input data: provide files as arguments or pipe data via stdin")
	}

	var tableNames []string

	if useStdin {
		name, err := loader.LoadStdin(stdinTableName)
		if err != nil {
			return err
		}
		tableNames = append(tableNames, name)
	}

	if len(cfg.filePaths) > 0 {
		names, err := loader.LoadFiles(cfg.filePaths)
		if err != nil {
			return err
		}
		tableNames = append(tableNames, names...)
	}

	cfg.tableNames = tableNames
	return nil
}

// execute runs either UI or CLI mode based on configuration.
// UI mode is used when query is empty, CLI mode when query is provided via -q flag.
func execute(database *db.DB, cfg *runConfig) error {
	if cfg.query == "" {
		result, err := ui.Run(database.DB, cfg.tableNames)
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
