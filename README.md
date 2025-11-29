# 🥑 qo

[![CI](https://github.com/kiki-ki/go-qo/actions/workflows/ci.yml/badge.svg)](https://github.com/kiki-ki/go-qo/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/kiki-ki/go-qo)](https://github.com/kiki-ki/go-qo/blob/main/LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/kiki-ki/go-qo)](https://goreportcard.com/report/github.com/kiki-ki/go-qo)

qo is a command-line tool for interactively querying JSON (... and soon, other formats) data using SQL.

![qo](https://github.com/user-attachments/assets/65aa3399-f8fe-473c-af8e-3548c70360ba)

## Install

<details>
<summary>Homebrew Tap</summary>

```bash
brew install kiki-ki/tap/go-qo
```

</details>

<details>
<summary>Shell Script</summary>

```sh
curl -sfL https://raw.githubusercontent.com/kiki-ki/go-qo/main/install.sh | sh
```

```sh
# specific directory
curl -sfL https://raw.githubusercontent.com/kiki-ki/go-qo/main/install.sh | BINDIR=./custom/bin sh
# specific version
curl -sfL https://raw.githubusercontent.com/kiki-ki/go-qo/main/install.sh | VERSION=v1.0.0 sh
```

</details>

<details>
<summary>Go Install</summary>

```sh
go install [github.com/kiki-ki/go-qo/cmd/qo@latest](https://github.com/kiki-ki/go-qo/cmd/qo@latest)
```

</details>

## Usage

```sh
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
