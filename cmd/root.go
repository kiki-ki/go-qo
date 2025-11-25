package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kiki-ki/go-qo/internal/db"
	"github.com/kiki-ki/go-qo/internal/parser"
	"github.com/kiki-ki/go-qo/internal/printer"
)

var rootCmd = &cobra.Command{
	Use:   "qo <file1> [file2...] <sql-query>",
	Short: "Execute SQL queries on JSON files",
	Args:  cobra.MinimumNArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[len(args)-1]
		filePaths := args[:len(args)-1]

		dbConn, err := db.NewDB()
		if err != nil {
			return fmt.Errorf("failed to init db: %w", err)
		}
		defer dbConn.Close()

		fmt.Fprintln(os.Stderr, "Loading tables...")

		for _, path := range filePaths {
			data, err := parser.ParseJSON(path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}

			tableName := db.SanitizeTableName(path)

			err = db.CreateTable(dbConn, tableName, data.Headers, data.Rows)
			if err != nil {
				return fmt.Errorf("failed to load table %s: %w", tableName, err)
			}

			fmt.Fprintf(os.Stderr, "  - %s => %s (%d rows)\n", path, tableName, len(data.Rows))
		}
		fmt.Fprintln(os.Stderr, "----------------")

		// Execute query
		rows, err := dbConn.Query(query)
		if err != nil {
			return fmt.Errorf("query execution failed: %w", err)
		}
		defer rows.Close()

		if err := printer.Print(rows); err != nil {
			return fmt.Errorf("failed to print results: %w", err)
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
