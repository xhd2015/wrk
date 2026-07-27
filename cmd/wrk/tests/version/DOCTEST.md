# wrk --version — embedded build version

## Version
0.0.3

Decision tree for root-level `wrk --version`: prints the embedded version from
`wrkcli/VERSION.txt` via `go:embed`. No git checkout is required.

**Layer:** L2 in-process CLI via `wrkcli.Capture` (short path — not binary e2e).

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
- **Run** — `wrkcli.Capture` with `Dir=RepoDir` and `WRK_HOME` env only (mutex
  serializes Capture within this suite).

## Tree Overview

```
version/
├── basic/
│   └── prints-version/           # wrk --version → stdout v0.0.1\n
├── help/
│   └── mentions-flag/            # wrk -h mentions --version
└── mutual-exclusion/
    └── with-list/                # wrk --version --list → non-zero
```

## How to Run

```sh
doctest vet ./cmd/wrk/tests/version
doctest test ./cmd/wrk/tests/version
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
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

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d // inject context available; Capture uses explicit Dir/Env only
	res := wrkcli.Capture(wrkcli.CaptureOpts{
		Args: append([]string(nil), req.Args...),
		Dir:  req.RepoDir,
		Env:  versionCaptureEnv(req),
	})
	return &Response{
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		ExitCode: res.ExitCode,
	}, nil
}
```
