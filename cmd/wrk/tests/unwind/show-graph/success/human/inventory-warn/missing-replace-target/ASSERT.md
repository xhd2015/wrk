## Expected Output

Stdout: graph banners + peel `.` (human dir identity).

Stderr: contains `warning:` (missing / non-git replace target).

## Expected

- Exit code 0 (do not fail show-graph solely for missing replace target).
- Stderr contains `warning:`.
- Human graph banners present; peel order includes `.`.
- Module dir `.` listed (single-repo identity).
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertShowGraphHumanBanners(t, resp.Stdout)
	assertShowGraphPeelOrderHuman(t, resp.Stdout, req.PeelOrder)
	assertModuleDirListed(t, resp.Stdout, ".")
	if !strings.Contains(resp.Stderr, "warning:") && !strings.Contains(resp.Stderr, "Warning:") {
		t.Fatalf("missing replace target must emit warning: on stderr; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
	}
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
