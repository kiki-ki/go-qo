package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	_ "modernc.org/sqlite"

	"github.com/kiki-ki/go-qo/internal/parser"
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

	createSQL := fmt.Sprintf("CREATE TABLE `%s` (%s)", tableName, strings.Join(colDefs, ", "))
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

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO `%s` VALUES (%s)", tableName, strings.Join(placeholders, ", "))

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for i, row := range rows {
		if _, err := stmt.Exec(row...); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to insert row %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// TableNameFromPath generates a table name from a file path.
//
// Anything SQLite would reject in an unquoted identifier becomes "_", so the
// name qo reports can be pasted straight into a query. Names are not unique
// across paths; callers loading several files resolve collisions themselves.
func TableNameFromPath(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// SQLite identifiers allow ASCII letters, digits and "_", plus any byte
	// above ASCII, which keeps non-ASCII filenames intact.
	name = strings.Map(func(r rune) rune {
		if r == '_' || r >= utf8.RuneSelf ||
			r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			return r
		}
		return '_'
	}, name)

	// An identifier can be neither empty nor digit-initial.
	if name == "" || name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}
	return name
}
