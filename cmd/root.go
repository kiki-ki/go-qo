// Package cmd provides the CLI commands for qo.
package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kiki-ki/go-qo/internal/cli"
	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/input"
	"github.com/kiki-ki/go-qo/internal/output"
	"github.com/kiki-ki/go-qo/internal/ui"
)

var version = "dev"

const stdinTableName = "tmp"

// options holds the flag values of a single invocation. Keeping them per
// command rather than in package state lets tests run independently.
type options struct {
	inputFormat  string
	outputFormat string
	query        string
	noHeader     bool
}

// newRootCmd builds the root command with its own flag state.
func newRootCmd() *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Version: version,
		Use:     "qo [files...]",
		Short:   "Query structured data with SQL",
		Long:    "qo is a TUI/CLI tool that lets you query structured data using SQL.",
		Example: strings.Join([]string{
			"  qo data.json                                        # Interactive TUI mode",
			"  cat data.json | qo                                  # Pipe to TUI, output to stdout",
			"  cat data.json | qo - other.json                     # Read stdin alongside files",
			`  qo -q "SELECT * FROM a JOIN b ..." a.csv b.json     # Mixed input formats`,
			`  qo -q "SELECT * FROM data" data.json                # Direct query mode`,
			`  qo -i csv -o json data.csv -q "SELECT * FROM data"  # CSV to JSON`,
		}, "\n"),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Silenced here rather than on the command so that flag parsing
			// errors, which happen before RunE, still show usage. Everything
			// past this point is a runtime failure that usage cannot help with.
			cmd.SilenceUsage = true
			return run(cmd, args, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.inputFormat, "input", "i", "json", "Input format: json, csv, tsv, psv (default: by file extension)")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "json", "Output format: json, jsonl, csv, tsv, psv, table")
	cmd.Flags().StringVarP(&opts.query, "query", "q", "", "SQL query to execute (if omitted, interactive mode)")
	cmd.Flags().BoolVar(&opts.noHeader, "no-header", false, "Treat first row as data, not header (CSV/TSV/PSV only)")

	return cmd
}

// runConfig holds the parsed configuration for a query run.
type runConfig struct {
	query      string
	filePaths  []string
	tableNames []string
}

func run(cmd *cobra.Command, args []string, opts *options) error {
	if err := validateFormats(opts); err != nil {
		return err
	}

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = database.Close() }()

	loader := input.NewLoader(database, input.Format(opts.inputFormat), &input.LoaderOptions{
		NoHeader: opts.noHeader,
		// An explicit -i applies to every file; otherwise each file's
		// extension decides, so mixed-format inputs can be joined.
		DetectFormat: !cmd.Flags().Changed("input"),
	})

	filePaths, stdinRequested := input.SplitArgs(args)
	useStdin, err := input.UseStdin(filePaths, stdinRequested)
	if err != nil {
		return err
	}

	cfg := &runConfig{
		query:     opts.query,
		filePaths: filePaths,
	}

	if err := loadData(loader, cfg, useStdin); err != nil {
		return err
	}

	return execute(database, cfg, opts, cmd.OutOrStdout())
}

// validateFormats checks if input/output formats are valid.
func validateFormats(opts *options) error {
	if !input.IsValidFormat(opts.inputFormat) {
		return fmt.Errorf("unsupported input format: %s (supported: %v)", opts.inputFormat, input.Formats())
	}
	if !output.IsValidFormat(opts.outputFormat) {
		return fmt.Errorf("unsupported output format: %s (supported: %v)", opts.outputFormat, output.Formats())
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
func execute(database *db.DB, cfg *runConfig, opts *options, out io.Writer) error {
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
		Format: output.Format(opts.outputFormat),
		Output: out,
	})
}

func Execute() error {
	return newRootCmd().Execute()
}
