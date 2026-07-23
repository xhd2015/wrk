# wrk --set-config — create UX defaults in config.json

## Version
0.0.3

Decision tree for `wrk --set-config`: management mode that merges create UX
defaults into `$WRK_HOME/config.json` and pretty-prints config with `--show`.

**Layer: L2 in-process CLI** — `Run` calls `wrkcli.RunCLI` (same dispatch as the
binary, captured writers). Not product-binary e2e.

# DSN (Domain Specific Notion)

- **wrk CLI** — `--set-config` is a standalone management mode (no git required
  for help/show/write). Mutually exclusive with other modes (`--list`, create
  positionals, `--no-config`, …).
- **config.json** — under `$WRK_HOME`; merge-only create UX keys on `--create`;
  preserve unknown top-level keys.
- **Help** — level-specific `-h`/`--help` for dispatcher, `--create`, and `--show`.
- **WRK_HOME** — isolated per test via `RunOptions.WrkHome` (no process Setenv).

## Tree Overview

```
set-config/
├── help/          # dispatcher / create / show usage
├── show/          # --show JSON
├── write/         # --create merge + negatives
└── mutual-exclusion/
```

## How to Run

```sh
doctest vet ./cmd/wrk/tests/set-config
doctest test ./cmd/wrk/tests/set-config
```

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot  string
	WrkHome   string
	RepoDir   string // effective cwd for in-process Run
	TargetDir string // optional first positional <dir>
	Args      []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	args := append([]string(nil), req.Args...)
	if req.TargetDir != "" {
		args = append([]string{req.TargetDir}, args...)
	}
	var stdout, stderr bytes.Buffer
	code := wrkcli.RunCLI(args, wrkcli.RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Dir:     req.RepoDir,
		WrkHome: req.WrkHome,
		WrkDate: wrkDate,
	})
	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: code,
	}, nil
}
```
