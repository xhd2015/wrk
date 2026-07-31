## Expected

- Non-zero exit.
- Stderr mentions linked worktree (and preferably `--pr`).
- No successful PR/show stdout (no bare URL success path, no create/attach tokens).
- Fake `gh` must **not** call `pr create` or `pr comment`.

## Errors

- Main-repo cwd is refused for bare `--pr` show (same linked-worktree gate as create).

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
		t.Fatalf("expected non-zero exit for bare --pr on main repo; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "linked worktree") && !strings.Contains(se, "not a linked") {
		t.Fatalf("stderr should mention linked worktree, got %q", resp.Stderr)
	}
	// Prefer naming --pr (soft if product reuses generic linked-worktree wording).
	_ = strings.Contains(resp.Stderr, "--pr")

	if strings.Contains(resp.Stdout, "PR created") || strings.Contains(resp.Stdout, "comment added") {
		t.Fatalf("must not claim PR success on main repo; stdout=%q", resp.Stdout)
	}

	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
	assertGhSubcmdNotCalled(t, invocs, "comment")
}
```
