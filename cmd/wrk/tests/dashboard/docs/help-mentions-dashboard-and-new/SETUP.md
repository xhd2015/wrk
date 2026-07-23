# Scenario

**Feature**: root help mentions dashboard and --new

```
wrk -h -> dashboard + --new
```

## Preconditions

- L2 in-process CLI via `wrkcli.RunCLI` (no product binary).
- Isolated `WRK_HOME` per leaf.

## Steps

1. Root `Setup` creates work root + WRK_HOME.
2. Leaves set `req.Args` for the help invocation.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	if len(req.Args) == 0 {
		req.Args = []string{"-h"}
	}
	return nil
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
}
```
