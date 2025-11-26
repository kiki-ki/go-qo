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

const stdinTableName = "t"

var (
	outputFormat string
	inputFormat  string
	verbose      bool
	queryFlag    string
)

var rootCmd = &cobra.Command{
	Use:   "qo [file1 file2...] [sql-query]",
	Short: "Execute SQL queries on JSON files",
	Long: `qo is a command-line tool that allows you to query JSON files using SQL.

Interactive Mode:
	qo data.json
	cat data.json | qo

CLI Mode:
	qo data.json "SELECT * FROM data"
	cat data.json | qo "SELECT * FROM t"

You can also provide the SQL via flag:
	qo -q "SELECT * FROM data" data.json
	cat data.json | qo -q "SELECT * FROM t"`,

	Args: cobra.ArbitraryArgs,
	RunE: runQuery,
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table (default) | json | csv")
	rootCmd.Flags().StringVarP(&inputFormat, "input", "i", "json", "Input format: json (default)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show informational messages")
	rootCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "SQL query to execute (if omitted, TUI mode)")
}

func runQuery(cmd *cobra.Command, args []string) error {
	if !input.IsValidFormat(inputFormat) {
		return fmt.Errorf("unsupported input format: %s (supported: %v)", inputFormat, input.Formats())
	}
	if !output.IsValidFormat(outputFormat) {
		return fmt.Errorf("unsupported output format: %s (supported: %v)", outputFormat, output.Formats())
	}

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	loader := input.NewLoader(database, input.Format(inputFormat), verbose)

	// Check if stdin has data
	stdinHasData, err := hasStdinData()
	if err != nil {
		return err
	}

	// --- 引数の解析とモード判定 ---
	var query string
	var filePaths []string
	var isTUI bool

	// Priority: --query/-q flag > positional args
	if queryFlag != "" {
		query = queryFlag
		filePaths = args
		if stdinHasData {
			if err := loader.LoadStdin(stdinTableName); err != nil {
				return err
			}
		}
	} else if stdinHasData {
		if len(args) > 0 {
			query = args[0]
		} else {
			isTUI = true
		}

		if err := loader.LoadStdin(stdinTableName); err != nil {
			return err
		}

	} else {
		if len(args) == 0 {
			isTUI = true
		} else if len(args) == 1 {
			filePaths = args
			isTUI = true
		} else {
			query = args[len(args)-1]
			filePaths = args[:len(args)-1]
		}
	}

	// Load files into database (if any)
	if len(filePaths) > 0 {
		if err := loader.LoadFiles(filePaths); err != nil {
			return err
		}
	}

	if isTUI {
		return tui.Run(database.DB)
	}
	return cli.Run(database.DB, query, &cli.Options{
		Format: output.Format(outputFormat),
		Output: os.Stdout,
	})
}

// hasStdinData checks if there's data available on stdin.
func hasStdinData() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to stat stdin: %w", err)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0, nil
}

func Execute() error {
	return rootCmd.Execute()
}
