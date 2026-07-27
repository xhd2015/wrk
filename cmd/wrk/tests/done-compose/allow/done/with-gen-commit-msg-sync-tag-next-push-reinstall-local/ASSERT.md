## Expected

- Flag layer accepts the full P3 ship combo:
  `--gen-commit-msg --commit --model=m --done --sync --tag-next --push --reinstall-local` (and `-y`).
- Stderr must **not** contain `mutually exclusive`.
- Stderr must **not** contain `--push is only valid with --tag-next`.
- Later-stage errors on main-repo cwd (e.g. not a linked worktree, no staged changes) are acceptable.

## Side Effects

- None required (flag-layer only).

## Exit Code

- Any, as long as failure is not flag-matrix rejection of this full combo.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("full pre+posts+reinstall ship still mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "--push is only valid with --tag-next") {
		t.Fatalf("push still tag-next-only with full P3 ship combo; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
