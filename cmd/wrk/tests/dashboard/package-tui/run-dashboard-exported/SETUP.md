# Scenario

**Feature**: `wrkcli/tui` exports `RunDashboard` plus supporting public types

```
# public API (locked for extraction)
package tui
  -> RunDashboard(opts RunDashboardOpts) error
  -> type Recipe
  -> type RunDashboardOpts

# discovery
go doc github.com/xhd2015/wrk/wrkcli/tui.RunDashboard
go doc github.com/xhd2015/wrk/wrkcli/tui.Recipe
go doc github.com/xhd2015/wrk/wrkcli/tui.RunDashboardOpts
  -> each resolves (exported symbols present)
```

## Steps

1. Ensure Go is on PATH; resolve module root.
2. Cheap `wrk -h` via root `Run`.
3. Assert package lists, then `go doc` for each required exported symbol.
4. Prefer doc text that shows `RunDashboard` is a **func** and Recipe / RunDashboardOpts are **types** (soft shape check; not full field inventory).

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
	_ = moduleRootFromDoctest(t, d)
	req.RepoDir = req.WorkRoot
	req.Args = []string{"-h"}
	return nil
}
```
