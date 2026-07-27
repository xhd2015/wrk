---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Package `github.com/xhd2015/wrk/wrkcli/tui` is listable.
- Direct imports of `tui` do **not** include parent `github.com/xhd2015/wrk/wrkcli` (exact import path match).
- Sibling imports such as `github.com/xhd2015/wrk/wrkcli/teapre` remain allowed when present.
- Parent package `github.com/xhd2015/wrk/wrkcli` remains listable (CLI package still exists).

## Errors

- Missing `tui` package → RED (same as importable).
- `tui` imports parent `wrkcli` → RED (cycle / wrong layering).

## Side Effects

- None beyond incidental `wrk -h`.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected Run error for wrk -h: %v", err)
	}
	assertPackageListed(t, wrkcliTuiImportPath)

	// Parent package must still resolve independently.
	parentOut, parentErr := goListInModule(t, wrkcliParentImportPath)
	if parentErr != nil {
		t.Fatalf("parent package %s must remain listable: %v\n%s",
			wrkcliParentImportPath, parentErr, parentOut)
	}

	importsOut, listErr := goListInModule(t, "-f", "{{join .Imports \"\\n\"}}", wrkcliTuiImportPath)
	if listErr != nil {
		t.Fatalf("go list imports for %s: %v\n%s", wrkcliTuiImportPath, listErr, importsOut)
	}
	for _, line := range strings.Split(importsOut, "\n") {
		imp := strings.TrimSpace(line)
		if imp == "" {
			continue
		}
		if imp == wrkcliParentImportPath {
			t.Fatalf("%s must not import parent package %s (use RunDashboardOpts callbacks); imports:\n%s",
				wrkcliTuiImportPath, wrkcliParentImportPath, importsOut)
		}
	}
}
```
