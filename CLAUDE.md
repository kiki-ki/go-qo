# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build        # build to ./bin/qo
make test         # run all tests
make lint         # go vet + golangci-lint
make check        # tidy, fmt, lint, vuln (run before committing)
make install-tool # install dev tools

go test ./internal/db/... -run TestFoo -v  # run a single test
```

## Architecture

Data flow: **stdin/files → `input.Loader` → `parser` → `db.DB` → `ui` or `cli` → `output.Printer` → stdout**

| Package | Role |
| --- | --- |
| `cmd/root.go` | Cobra entry point. Switches between TUI and CLI based on the `-q` flag |
| `internal/input` | Reads stdin/files and dispatches to parser. Stdin always maps to table `"tmp"`; files use a sanitized filename |
| `internal/parser` | Converts bytes to `ParsedData` (columns + rows). Parsers self-register via `init()` |
| `internal/db` | In-memory SQLite via `modernc.org/sqlite` (CGO-free) |
| `internal/ui` | Bubble Tea TUI. `TableState` implements custom horizontal column scrolling |
| `internal/cli` | Non-interactive execution when `-q` is provided |
| `internal/output` | Formats `*sql.Rows` as JSON, JSONL, CSV, TSV, PSV, or table |

**Key design notes:**
- TUI opens `/dev/tty` directly so it renders even when stdin is a pipe
- On Esc, the TUI passes the final query to `cli.Run()` for stdout output, enabling interactive-then-pipe workflows
- `goimports` local prefix: `github.com/kiki-ki/go-qo`
