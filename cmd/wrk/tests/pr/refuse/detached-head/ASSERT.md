## Expected

- Non-zero exit.
- Stderr mentions detached HEAD (case-insensitive ok).
- No PR success tokens on stdout.
- Fake `gh` create/comment not called.

## Errors

- Detached HEAD cannot be used as PR head branch.

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
		t.Fatalf("expected non-zero for detached HEAD; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "detached") {
		t.Fatalf("stderr should mention detached HEAD, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "PR created") || strings.Contains(resp.Stdout, "comment added") {
		t.Fatalf("must not claim PR success; stdout=%q", resp.Stdout)
	}
	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
