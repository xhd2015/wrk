## Expected

- Flag layer accepts `--done` + `--tag-next` (no mutual-exclusion error).
- Stderr must **not** contain `mutually exclusive`.
- Exit may be non-zero for later-stage reasons on a main-repo cwd (e.g. `not a linked worktree`) — that still counts as flag-layer pass.
- Exit 0 is also acceptable if a future full pipeline succeeds.

## Side Effects

- None required for P1 (no tag/merge success asserted).

## Exit Code

- Any (0 or non-zero), as long as the failure is not flag mutual exclusion.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --done --tag-next as mutually exclusive; stderr=%q stdout=%q exit=%d",
			se, resp.Stdout, resp.ExitCode)
	}
}
```
