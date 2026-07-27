# Scenario

**Feature**: `wrkcli/tui` avoids import cycle with parent `wrkcli`

```
# dependency direction
wrkcli  -->  wrkcli/tui   (allowed: parent may import child)
wrkcli/tui  -/->  wrkcli  (forbidden: child must not import parent)

# inject deps instead of importing parent
RunDashboardOpts carries callbacks:
  HasAddableDirt, IsMainCheckout, GitAddAll, RunCompose, ComposeArgv
```

## Steps

1. Ensure Go is on PATH; resolve module root.
2. Cheap `wrk -h` via root `Run`.
3. Require package listable.
4. `go list -f '{{join .Imports "\n"}}'` on tui — no line equals `github.com/xhd2015/wrk/wrkcli`.
5. Soft: `go list` parent `github.com/xhd2015/wrk/wrkcli` still succeeds (module still builds the CLI package).

```go
import (
	"os/exec"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	_ = moduleRootFromDoctest(t)
	req.RepoDir = req.WorkRoot
	req.Args = []string{"-h"}
	return nil
}
```
