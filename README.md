# qo

A command-line tool to query JSON data using SQL.

## Installation

```bash
go install github.com/kiki-ki/go-qo/cmd/qo@latest
```

## Usage

```bash
# Interactive mode (default)
qo data.json                              # Open interactive editor
cat data.json | qo                        # Pipe data to interactive editor

# CLI mode
qo -q "SELECT * FROM data" data.json      # Direct query execution
cat data.json | qo -q "SELECT * FROM tmp" # Query piped data
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input format: `json` (default) |
| `--output` | `-o` | Output format: `table` (default), `json`, `csv` |
| `--query` | `-q` | SQL query (skips interactive mode) |

## Interactive mode

| Key | Mode | Action |
| - | - | - |
| `Tab` | ALL | Switch between Query/Table mode |
| `Esc/Ctrl+C` | ALL | Quit |
| `Enter` | "QUERY" | Execute query and exit |
| `↑/↓` or `j/k` | "TABLE" | Scroll rows |
| `←/→` or `h/l` | "TABLE" | Scroll columns |

## License

MIT
