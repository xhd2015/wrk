# Scenario

**Feature**: package `github.com/xhd2015/wrk/wrkcli/tui` exists in the wrk module

```
# go list
module github.com/xhd2015/wrk
  -> go list github.com/xhd2015/wrk/wrkcli/tui
  -> prints the import path (package present under wrkcli/tui/)
```

## Steps

1. Ensure Go is on PATH; resolve module root (must contain `go.mod`).
2. Run cheap `wrk -h` via root `Run` (harness only).
3. Assert via `go list` that `github.com/xhd2015/wrk/wrkcli/tui` is a package in this module.

```go
import (
	"github.com/xhd2015/doctest/session"
	"os/exec"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Package existence check needs go list against the wrk module.
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	req.RepoDir = req.WorkRoot
	req.Args = []string{"-h"}
	return nil
}
```
