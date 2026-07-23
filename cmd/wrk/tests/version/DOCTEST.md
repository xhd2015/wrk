# wrk --version — embedded build version

## Version
0.0.3

Decision tree for root-level `wrk --version`: prints the embedded version from
`wrkcli/VERSION.txt` via `go:embed`. No git checkout is required.

**Layer: L2 in-process CLI** — `Run` calls `wrkcli.RunCLI` (same dispatch as the
binary, captured writers). Not product-binary e2e.

# DSN (Domain Specific Notion)

- **wrk CLI** — top-level `--version` is a standalone action flag; it must be
  the only mode/behavior flag on the command line (mutually exclusive with
  `--list`, `--done`, create, `skill`, etc.).
- **Embedded VERSION.txt** — `wrkcli/VERSION.txt` holds `0.0.1` (no `v` prefix);
  the binary prints `v0.0.1` plus trailing newline on stdout.
- **wrk --version** — stdout `v0.0.1\n`, stderr empty, exit 0; does not append
  `events.jsonl`.
- **wrk -h / wrk --help** — root usage documents `--version`.
- **wrk --version + other flag** — non-zero exit; stderr mentions mutual
  exclusion or unexpected arguments; stdout empty.
- **WRK_HOME** — isolated per test at `{WorkRoot}/.wrk`; version action does
  not write events.

## Tree Overview

```
version/
├── basic/
│   └── prints-version/           # wrk --version → stdout v0.0.1\n
├── help/
│   └── mentions-flag/            # wrk -h contains --version
└── mutual-exclusion/
    └── with-list/                # wrk --version --list → error
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| V1 | basic/prints-version | `wrk --version` → exit 0; stdout `v0.0.1\n`; no events.jsonl |
| V2 | help/mentions-flag | `wrk -h` → exit 0; help mentions `--version` |
| V3 | mutual-exclusion/with-list | `wrk --version --list` → exit ≠0; mutually exclusive |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/version
doctest test ./cmd/wrk/tests/version
doctest test ./cmd/wrk/tests/version/basic/prints-version
doctest test ./cmd/wrk/tests/version/help/mentions-flag
doctest test ./cmd/wrk/tests/version/mutual-exclusion/with-list
```

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process effective cwd for in-process Run
	Args     []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var stdout, stderr bytes.Buffer
	code := wrkcli.RunCLI(req.Args, wrkcli.RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Dir:     req.RepoDir,
		WrkHome: req.WrkHome,
	})
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
	}, nil
}
```
