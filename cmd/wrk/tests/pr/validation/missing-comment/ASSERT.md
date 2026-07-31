## Expected

- Non-zero exit.
- Stderr mentions `--comment` (required / missing for incomplete create).
- No PR success tokens; `gh pr create` not called.

## Errors

- Create mode requires both `--title` and `--comment`. Bare `--pr` is show (P1); `--pr --comment` alone is comment-only (P2). Title without comment remains incomplete create.

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
		t.Fatalf("expected non-zero without --comment; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "--comment") {
		t.Fatalf("stderr should mention --comment, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "PR created") {
		t.Fatalf("must not create PR; stdout=%q", resp.Stdout)
	}
	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
}
```
