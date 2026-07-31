## Expected

- Non-zero exit.
- Stderr mentions `--comment` (empty after trim / required).
- No `comment added` / `PR created` on stdout.
- Fake `gh`: **`pr create` and `pr comment` not called**.

## Errors

- Comment-only still requires non-empty body after trim (same empty rule as create path; create-path empty covered under `validation/empty-comment/`).

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
		t.Fatalf("expected non-zero for empty comment-only; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "--comment") {
		t.Fatalf("stderr should mention --comment, got %q", resp.Stderr)
	}
	for _, tok := range []string{"comment added", "PR created"} {
		if strings.Contains(resp.Stdout, tok) {
			t.Fatalf("must not print %q; stdout=%q", tok, resp.Stdout)
		}
	}
	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
