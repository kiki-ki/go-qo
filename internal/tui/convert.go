package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
)

// sqlRowsToTable : SQLの結果をBubble Teaのテーブルデータに変換する
func sqlRowsToTable(rows *sql.Rows) ([]table.Column, []table.Row, error) {
	// 1. カラム情報の取得
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	// table.Column の作成
	// 幅(Width)は一旦固定か、あるいは文字数に合わせて計算もできますが、まずは簡易的に設定
	tCols := make([]table.Column, len(cols))
	for i, c := range cols {
		tCols[i] = table.Column{Title: c, Width: 15} // 幅は適当(15)
	}

	// 2. データの取得
	var tRows []table.Row

	// Scan用の準備
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, nil, err
		}

		// 1行分のデータ (文字列の配列)
		rowData := make(table.Row, len(cols))
		for i, val := range values {
			if val == nil {
				rowData[i] = "(NULL)"
			} else {
				switch v := val.(type) {
				case []byte:
					rowData[i] = string(v)
				case float64:
					// 1.00 -> 1
					if float64(int64(v)) == v {
						rowData[i] = fmt.Sprintf("%d", int64(v))
					} else {
						rowData[i] = fmt.Sprintf("%g", v)
					}
				default:
					rowData[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		tRows = append(tRows, rowData)
	}

	return tCols, tRows, nil
}
