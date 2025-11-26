package tui

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// スタイル定義 (Lipgloss)
var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// ヘッダーや選択行のスタイルもここで定義できます
)

type Model struct {
	db        *sql.DB
	table     table.Model     // テーブルコンポーネント
	textInput textinput.Model // 入力コンポーネント
	err       error           // SQLエラー用
}

func NewModel(db *sql.DB) Model {
	// 1. 入力欄の初期化
	ti := textinput.New()
	ti.Placeholder = "SQL Query..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 100

	// 2. テーブルの初期化
	t := table.New(
		table.WithColumns([]table.Column{{Title: "Results", Width: 20}}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false), // 最初はフォーカスしない(入力欄優先)
		table.WithHeight(10),     // 表示する行数
	)

	// テーブルの見た目カスタマイズ
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		db:        db,
		table:     t,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		// Tabキーでフォーカス切り替え (入力欄 <-> テーブル)
		case tea.KeyTab:
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.table.Focus()
			} else {
				m.table.Blur()
				m.textInput.Focus()
			}
		}
	}

	// コンポーネントのアップデート
	m.table, _ = m.table.Update(msg)
	m.textInput, cmd = m.textInput.Update(msg)

	// ★リアルタイムSQL実行ロジック
	// 入力欄にフォーカスがある時だけ実行を試みる
	if m.textInput.Focused() {
		query := m.textInput.Value()
		// クエリが空なら何もしない
		if query != "" {
			rows, err := m.db.Query(query)
			if err == nil {
				// 成功したらテーブルを更新
				defer rows.Close()
				cols, tableRows, err := sqlRowsToTable(rows)
				if err == nil {
					m.table.SetColumns(cols)
					m.table.SetRows(tableRows)
					m.err = nil // エラーなし
				}
			} else {
				// SQL構文エラーなどは画面に出してもいいが、
				// 入力中は頻発するので一旦内部保持のみにするか、
				// ステータスバーに出すのが一般的
				m.err = err
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	// エラー表示エリア
	errView := ""
	if m.err != nil {
		errView = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("\nError: %v", m.err))
	}

	return baseStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			" SQL Editor (Tab to switch focus)",
			m.textInput.View(),
			errView,
			"\n",
			m.table.View(), // テーブル表示
		),
	) + "\n"
}

func Run(db *sql.DB) error {
	p := tea.NewProgram(NewModel(db), tea.WithAltScreen()) // AltScreen: 全画面モード
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
