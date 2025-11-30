# 🥢 qo

[![CI](https://github.com/kiki-ki/go-qo/actions/workflows/ci.yml/badge.svg)](https://github.com/kiki-ki/go-qo/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/kiki-ki/go-qo)](https://github.com/kiki-ki/go-qo/blob/main/LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/kiki-ki/go-qo)](https://goreportcard.com/report/github.com/kiki-ki/go-qo)

> **qo** [cue-oh] *noun.*
>
> 1. Abbreviation for **"Query & Out"**.
> 2. The peace of mind obtained by filtering JSON streams with SQL instead of complex path syntax.

**qo** is a minimalist TUI that lets you query JSON and CSV files using SQL.
Pick what you need with SQL, and get **Out** to the pipeline.

![qo demo](https://github.com/user-attachments/assets/65aa3399-f8fe-473c-af8e-3548c70360ba)

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
qo -q "SELECT * FROM x JOIN y ON x.id = y.x_id" x.json y.json
```

### CSV Support

```bash
# CSV input
qo -i csv data.csv
qo -i csv -q "SELECT name, age FROM data WHERE age > 30" data.csv

# CSV without header row
qo -i csv --no-header data.csv
# Columns are named: col1, col2, col3, ...
```

### Filter JSON API Response

Interactively filter JSON response from any API.

```bash
curl -s https://api.github.com/repos/kiki-ki/go-qo/commits | qo
```

## Options

| Flag | Short | Description |
| :--- | :--- | :--- |
| `--input` | `-i` | Input format: `json` (default), `csv` |
| `--output` | `-o` | Output format: `table` (default), `json`, `csv` |
| `--query` | `-q` | Run SQL query directly (Skip TUI) |
| `--no-header` | | Treat first row as data, not header (CSV only) |

## UI Controls

| Key | Mode | Action |
| :--- | :--- | :--- |
| `Tab` | ALL | Switch between Query/Table mode |
| `Esc` / `Ctrl+C` | ALL | Quit (Output nothing) |
| `Enter` | QUERY | **Output result to stdout** and Exit |
| `↑` `↓` / `j` `k` | TABLE | Scroll rows |
| `←` `→` / `h` `l` | TABLE | Scroll columns |

## Roadmap

* [x] JSON Support
* [x] CSV Support
* [ ] TSV Support

## License

MIT
