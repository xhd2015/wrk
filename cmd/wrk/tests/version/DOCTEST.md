# wrk --version — embedded build version

## Version
0.0.2

Decision tree for root-level `wrk --version`: prints the embedded version from
`wrkcli/VERSION.txt` via `go:embed`. No git checkout is required.

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
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process cwd when running wrk
	Args     []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)

	args := append([]string(nil), req.Args...)
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = versionWrkEnv(req)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```