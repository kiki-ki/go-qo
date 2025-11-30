# 🥢 qo

[![CI](https://github.com/kiki-ki/go-qo/actions/workflows/ci.yml/badge.svg)](https://github.com/kiki-ki/go-qo/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/kiki-ki/go-qo)](https://github.com/kiki-ki/go-qo/blob/main/LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/kiki-ki/go-qo)](https://goreportcard.com/report/github.com/kiki-ki/go-qo)

> **qo** [cue-oh] *noun.*
>
> 1. Abbreviation for **"Query & Out"**.
> 2. The peace of mind obtained by filtering data with SQL instead of complex syntax.

**qo** is a minimalist TUI that lets you query JSON, CSV, and TSV files using SQL.
**Query** what you need, and get it **Out** to the pipeline.

![qo demo](doc/demo/demo.gif)

## Why qo?

* **Muscle Memory**: Use the SQL syntax you've known for years.
* **Pipeline Native**: Reads from `stdin`, writes to `stdout`.
* **Interactive**: Don't guess the query. See the result, then hit Enter.

## Install

### Homebrew (macOS/Linux)

```bash
brew install kiki-ki/tap/qo
```

### Shell Script

```bash
curl -sfL https://raw.githubusercontent.com/kiki-ki/go-qo/main/install.sh | sh
```

### Go Install

```bash
go install github.com/kiki-ki/go-qo/cmd/qo@latest
```

## Usage

**qo** reads from both file arguments and standard input (stdin).

```bash
# Interactive mode (Open TUI)
cat x.json | qo
qo x.json y.json

# Non-interactive mode (Direct output)
cat x.json | qo -q "SELECT * FROM tmp WHERE id > 100"
qo -q "SELECT * FROM x JOIN y ON x.id = y.x_id" x.json y.json"
```

### Pipe-Friendly TUI

TUI mode works seamlessly with pipes. Explore data interactively, then pass the result to other tools.

```bash
# Fetch JSON API > Filter interactively with qo > Format with jq
curl -s https://api.github.com/repos/kiki-ki/go-qo/commits | qo | jq '.[].sha'

# Explore > Filter > Compress
cat large.json | qo | gzip > filtered.json.gz
```

### Query Logs & Aggregate

Use SQL to analyze structured data.

```bash
# Filter error logs
cat app_json.log | qo -q "SELECT timestamp, message FROM tmp WHERE level = 'error'"

# Aggregate sales by region
qo -i csv sales.csv -o csv -q "SELECT region, SUM(amount) FROM sales GROUP BY region"
```

### Convert Formats

Transform between JSON, CSV, and TSV.

```bash
qo -o csv data.json -q "SELECT id, name FROM data"             # JSON → CSV
qo -i csv -o json users.csv -q "SELECT * FROM users"           # CSV → JSON
qo -i csv --no-header raw.csv -q "SELECT col1, col2 FROM raw"  # Headerless CSV
```

## Options

| Flag | Short | Description |
| :--- | :--- | :--- |
| `--input` | `-i` | Input format: `json` (default), `csv`, `tsv` |
| `--output` | `-o` | Output format: `json` (default), `table`, `csv`, `tsv` |
| `--query` | `-q` | Run SQL query directly (Skip TUI) |
| `--no-header` | | Treat first row as data, not header (CSV/TSV only) |

## UI Controls

| Key | Mode | Action |
| :--- | :--- | :--- |
| `Tab` | ALL | Switch between Query/Table mode |
| `Esc` / `Ctrl+C` | ALL | Quit (Output nothing) |
| `Enter` | QUERY | **Output result to stdout** and Exit |
| `↑` `↓` / `j` `k` | TABLE | Scroll rows |
| `←` `→` / `h` `l` | TABLE | Scroll columns |

## License

MIT
