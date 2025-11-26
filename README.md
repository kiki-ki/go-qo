# qo

A command-line tool to query JSON data using SQL with an interactive editor.

## Installation

```bash
go install github.com/kiki-ki/go-qo/cmd/qo@latest
```

## Usage

```bash
# TUI usage
qo -i json data.json                              # Open TUI editor with file data
cat data.json | qo -i json                        # Open TUI editor with stdin data
# CLI usage
qo -i json -q "SELECT * FROM data" data.json      # Direct query to file data
cat data.json | qo -i json -q "SELECT * FROM tmp" # Direct query to stdin data
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input format: `json` (default) |
| `--output` | `-o` | Output format: `table` (default), `json`, `csv` |
| `--query` | `-q` | SQL query (enables CLI mode) |

## TUI editor (default)

| Key | Mode | Action |
| - | - | - |
| `Tab` | ALL | Switch between Query/Table mode |
| `Esc/Ctrl+C` | ALL | Quit |
| `Enter` | "QUERY" | Execute query and exit |
| `↑/↓` or `j/k` | "TABLE" | Scroll rows |
| `←/→` or `h/l` | "TABLE" | Scroll columns |

## License

MIT
