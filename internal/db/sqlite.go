// Package db provides SQLite database operations for querying parsed data.
package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kiki-ki/go-qo/internal/parser"
	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with additional functionality.
type DB struct {
	*sql.DB
}

// New creates a new in-memory SQLite database.
func New() (*DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and optimize for in-memory usage
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &DB{DB: db}, nil
}

// LoadData loads parsed data into a table.
func (db *DB) LoadData(tableName string, data *parser.ParsedData) error {
	if err := db.createTable(tableName, data.Columns); err != nil {
		return err
	}
	return db.insertRows(tableName, data.Columns, data.Rows)
}

// createTable creates a table with the given columns.
func (db *DB) createTable(tableName string, columns []parser.Column) error {
	colDefs := make([]string, len(columns))
	for i, col := range columns {
		colDefs[i] = fmt.Sprintf("`%s` %s", col.Name, col.Type.String())
	}

	createSQL := fmt.Sprintf(
		"CREATE TABLE `%s` (%s)",
		tableName,
		strings.Join(colDefs, ", "),
	)

	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}
	return nil
}

// insertRows inserts all rows into the table using a transaction.
func (db *DB) insertRows(tableName string, columns []parser.Column, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf(
		"INSERT INTO `%s` VALUES (%s)",
		tableName,
		strings.Join(placeholders, ", "),
	)

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range rows {
		if _, err := stmt.Exec(row...); err != nil {
			return fmt.Errorf("failed to insert row %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// TableNameFromPath generates a valid table name from a file path.
// e.g., "data/users.json" -> "users"
func TableNameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Replace invalid characters with underscores
	replacer := strings.NewReplacer(
		"-", "_",
		" ", "_",
		".", "_",
	)
	return replacer.Replace(name)
}

// NewDB is deprecated, use New instead.
// Deprecated: Use New() instead.
func NewDB() (*sql.DB, error) {
	return sql.Open("sqlite", ":memory:")
}

// CreateTable is deprecated, use DB.LoadData instead.
// Deprecated: Use DB.LoadData() instead.
func CreateTable(db *sql.DB, tableName string, headers []string, rows [][]interface{}) error {
	colsDef := make([]string, len(headers))
	for i, h := range headers {
		colsDef[i] = fmt.Sprintf("`%s` TEXT", h)
	}
	createSQL := fmt.Sprintf("CREATE TABLE `%s` (%s)", tableName, strings.Join(colsDef, ", "))

	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("create table error: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO `%s` VALUES (%s)", tableName, strings.Join(placeholders, ", "))

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.Exec(row...); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert error: %w", err)
		}
	}

	return tx.Commit()
}

// SanitizeTableName is deprecated, use TableNameFromPath instead.
// Deprecated: Use TableNameFromPath() instead.
func SanitizeTableName(path string) string {
	return TableNameFromPath(path)
}
