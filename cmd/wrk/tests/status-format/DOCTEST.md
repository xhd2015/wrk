# wrk --status — Status line format wording (`clean` / `dirty (… added…)`)

## Version
0.0.3

Focused regression tree for the **Status:** field wording owned by wrk
(`clean` vs `dirty (N staged, N changed, N renamed, N deleted, N untracked)`). Not a full
`--status` layout suite — only the format contract for clean and dirty porcelain
fixtures via L2 `wrk --status`.

**Layer:** L2 in-process CLI via `wrkcli.Capture` (short path — not binary e2e).

# DSN (Domain Specific Notion)

- **wrk CLI** — `wrk --status` from a git checkout root prints one status block.
- **Status format (wrk-owned)** — value is exactly `clean` when porcelain is
  empty; otherwise
  `dirty (<staged> staged, <changed> changed, <renamed> renamed, <deleted> deleted, <untracked> untracked)`
  (all five buckets always present; no ANSI in this tree).
- **Wrk taxonomy** — porcelain `??` → **untracked**; index `A`/`AM` → **staged**
  (path-once; `AM` does not also count as changed); `M`/default → **changed**.
- **Main-repo block** — from checkout root: `Dir: .`, `Branch`, `Commit`,
  `Status`, `Remote: (no upstream)` (no upstream in fixtures).
- **WRK_HOME** — isolated per leaf at `{WorkRoot}/.wrk`.
- **Run** — `wrkcli.Capture` with `Dir=RepoDir`, `WRK_HOME` + `WRK_DATE` env.

## Tree Overview

```
status-format/
├── clean/                 # empty porcelain → Status: clean
├── dirty-untracked/       # one ?? → dirty (… 1 untracked)
├── dirty-staged-added/    # staged A → dirty (1 staged, … 0 untracked)
└── dirty-am/              # AM path-once → 1 staged, 0 changed
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | clean | Clean committed checkout → `Status: clean` |
| 2 | dirty-untracked | Untracked file → `… 1 untracked` |
| 3 | dirty-staged-added | Staged new file → `1 added … 0 untracked` |
| 4 | dirty-am | Staged-new then edited → `1 staged, 0 changed …` |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/status-format
doctest test ./cmd/wrk/tests/status-format
doctest test ./cmd/wrk/tests/status-format/clean
doctest test ./cmd/wrk/tests/status-format/dirty-untracked
doctest test ./cmd/wrk/tests/status-format/dirty-staged-added
doctest test ./cmd/wrk/tests/status-format/dirty-am
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
