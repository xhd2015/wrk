# wrk --status — Status line format wording (`clean` / `dirty (… added…)`)

## Version
0.0.2

Focused regression tree for the **Status:** field wording owned by wrk
(`clean` vs `dirty (N added, N changed, N renamed, N deleted)`). Not a full
`--status` layout suite — only the format contract for clean and dirty porcelain
fixtures via L2 `wrk --status`.

**Layer:** L2 in-process CLI via `wrkcli.Capture` (short path — not binary e2e).

# DSN (Domain Specific Notion)

- **wrk CLI** — `wrk --status` from a git checkout root prints one status block.
- **Status format (wrk-owned)** — value is exactly `clean` when porcelain is
  empty; otherwise
  `dirty (<added> added, <changed> changed, <renamed> renamed, <deleted> deleted)`
  (all four buckets always present; no ANSI in this tree).
- **Wrk taxonomy** — porcelain `??` untracked counts as **added** (same as
  index `A` / `wrk --projects`).
- **Main-repo block** — from checkout root: `Dir: .`, `Branch`, `Commit`,
  `Status`, `Remote: (no upstream)` (no upstream in fixtures).
- **WRK_HOME** — isolated per leaf at `{WorkRoot}/.wrk`.
- **Run** — `wrkcli.Capture` with `Dir=RepoDir`, `WRK_HOME` + `WRK_DATE` env.

## Tree Overview

```
status-format/
├── clean/                 # empty porcelain → Status: clean
└── dirty-added/           # one ?? untracked → dirty (1 added, 0 changed, 0 renamed, 0 deleted)
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | clean | Clean committed checkout → `Status: clean` |
| 2 | dirty-added | Untracked file → `Status: dirty (1 added, 0 changed, 0 renamed, 0 deleted)` |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/status-format
doctest test ./cmd/wrk/tests/status-format
doctest test ./cmd/wrk/tests/status-format/clean
doctest test ./cmd/wrk/tests/status-format/dirty-added
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
	RepoDir  string // process cwd when running wrk --status
	MainRepo string
	Args     []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	args := append([]string(nil), req.Args...)
	if len(args) == 0 {
		args = []string{"--status"}
	}
	res := wrkcli.Capture(wrkcli.CaptureOpts{
		Args: args,
		Dir:  req.RepoDir,
		Env:  statusFormatCaptureEnv(req),
	})
	return &Response{
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		ExitCode: res.ExitCode,
	}, nil
}
```
