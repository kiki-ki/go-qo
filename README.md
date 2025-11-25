# qo

A command-line tool to query JSON object(s) using SQL.

## Features

TODO

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
qo [file1 file2...] <sql-query> [flags]
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Output format: `table` (default), `json`, `csv` |
| `--verbose` | `-v` | Show informational messages |

## Examples

```bash
# JSON file / Table name is decided by filename
qo ./data/users.json "SELECT name, age FROM users WHERE age > 30"
qo users.json tests.json "SELECT u.name, t.point FROM users u JOIN tests t ON u.id = t.user_id"

# Stdin (Pipe) / Temporary table name is `t`.
curl -s https://api.example.com/users | qo "SELECT * FROM t WHERE age > 30"
echo '[{"id":1},{"id":2}]' | qo "SELECT * FROM t"
```

## License

MIT
