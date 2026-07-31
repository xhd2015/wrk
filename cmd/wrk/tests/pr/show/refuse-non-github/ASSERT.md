## Expected

- Non-zero exit.
- Stderr mentions github.com and/or origin (host requirement).
- No `PR created` / `comment added` on stdout.
- Fake `gh` create/comment not called.

## Errors

- Only github.com origin is supported for bare `--pr` show (same gate as create).

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for bare --pr with non-github origin; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "github") && !strings.Contains(se, "origin") {
		t.Fatalf("stderr should explain github/origin requirement, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "PR created") || strings.Contains(resp.Stdout, "comment added") {
		t.Fatalf("must not claim PR success; stdout=%q", resp.Stdout)
	}
	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
