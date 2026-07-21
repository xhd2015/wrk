## Expected

- Package `github.com/xhd2015/wrk/wrkcli/tui` is listable.
- Exported symbols exist and are documentable via `go doc`:
  - `RunDashboard` — function entry for the interactive dashboard TUI
  - `Recipe` — public recipe/struct type for compose stages
  - `RunDashboardOpts` — public options/struct type (workdir, status, injectable callbacks)
- Doc output for `RunDashboard` includes the name `RunDashboard` and looks like a function (contains `func`).
- Doc output for `Recipe` and `RunDashboardOpts` includes each type name (and preferably `type`).

## Errors

- Missing package or missing export → `go list` / `go doc` fails → RED until implementer wires public API.

## Side Effects

- None on disk; no TUI process launched beyond incidental `wrk -h`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected Run error for wrk -h: %v", err)
	}
	assertPackageListed(t, d, wrkcliTuiImportPath)

	// RunDashboard must be an exported function.
	doc, docErr := goDocInModule(t, d, wrkcliTuiImportPath+".RunDashboard")
	if docErr != nil {
		t.Fatalf("tui must export RunDashboard (go doc): %v\n%s", docErr, doc)
	}
	if !strings.Contains(doc, "RunDashboard") {
		t.Fatalf("go doc RunDashboard missing name; got %q", doc)
	}
	if !strings.Contains(doc, "func") {
		t.Fatalf("RunDashboard should document as a function (func …); got %q", doc)
	}

	// Recipe type.
	doc, docErr = goDocInModule(t, d, wrkcliTuiImportPath+".Recipe")
	if docErr != nil {
		t.Fatalf("tui must export type Recipe (go doc): %v\n%s", docErr, doc)
	}
	if !strings.Contains(doc, "Recipe") {
		t.Fatalf("go doc Recipe missing name; got %q", doc)
	}

	// RunDashboardOpts type.
	doc, docErr = goDocInModule(t, d, wrkcliTuiImportPath+".RunDashboardOpts")
	if docErr != nil {
		t.Fatalf("tui must export type RunDashboardOpts (go doc): %v\n%s", docErr, doc)
	}
	if !strings.Contains(doc, "RunDashboardOpts") {
		t.Fatalf("go doc RunDashboardOpts missing name; got %q", doc)
	}
}
```
