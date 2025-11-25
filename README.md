# qo

A command-line tool to query JSON files using SQL.

## Features

- Query JSON files using standard SQL syntax
- Support for multiple JSON files with JOIN operations
- Automatic type inference (INTEGER, REAL, TEXT, BOOLEAN)
- Multiple output formats: table, JSON, CSV
- In-memory SQLite database for fast queries

## Installation

```bash
go install github.com/kiki-ki/go-qo/cmd/qo@latest
```

Or build from source:

```bash
git clone https://github.com/kiki-ki/go-qo.git
cd go-qo
go build -o qo ./cmd/qo
```

## Usage

```bash
qo <file1> [file2...] <sql-query> [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Output format: `table` (default), `json`, `csv` |
| `--verbose` | `-v` | Show informational messages |

### Examples

Basic query:

```bash
qo data.json "SELECT * FROM data"
```

Filter and select specific columns:

```bash
qo users.json "SELECT name, age FROM users WHERE age > 30"
```

JOIN multiple files:

```bash
qo users.json orders.json "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id"
```

Output as JSON:

```bash
qo data.json "SELECT * FROM data" --format json
```

Output as CSV:

```bash
qo data.json "SELECT * FROM data" -f csv > output.csv
```

Verbose mode (show table loading info):

```bash
qo data.json "SELECT * FROM data" -v
```

## Table Naming

The table name is derived from the filename:

- `data.json` → `data`
- `my-users.json` → `my_users`
- `path/to/orders.json` → `orders`

## Type Inference

The parser automatically infers column types from JSON values:

| JSON Type | SQL Type |
|-----------|----------|
| String | TEXT |
| Integer | INTEGER |
| Float | REAL |
| Boolean | INTEGER (0/1) |
| Object/Array | TEXT (raw JSON) |
| Null | NULL |

When values have mixed types, the parser automatically widens to the most compatible type.

## License

MIT
