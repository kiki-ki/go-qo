package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func NewDB() (*sql.DB, error) {
	return sql.Open("sqlite", ":memory:")
}

// Create table and insert data.
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

// Decide table name from file path.
// e.g., "data/users.json" -> "users"
func SanitizeTableName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(name)
}
