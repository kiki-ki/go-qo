# 🥑 qo

A command-line tool for interactively querying JSON (... and soon, else formats) data using SQL.

![qo](https://github.com/user-attachments/assets/65aa3399-f8fe-473c-af8e-3548c70360ba)

## Install

(Homebrew Tap)

```bash
brew install kiki-ki/tap/go-qo
```

(go install)

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

### Interactive mode usage

| Key | Mode | Action |
| - | - | - |
| `Tab` | ALL | Switch between Query/Table mode |
| `Esc/Ctrl+C` | ALL | Quit |
| `Enter` | QUERY | Execute query and exit |
| `↑↓` or `jk` | TABLE | Scroll rows |
| `←→` or `hl` | TABLE | Scroll columns |

## License

MIT
