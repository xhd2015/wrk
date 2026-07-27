
## Expected

- Non-zero exit (replace guard after cascade).
- Dep fix merged into dep main.
- External dependency worktree removed.
- Consumer linked worktree still exists.

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
		t.Fatalf("expected non-zero exit (replace guard after cascade), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	lower := strings.ToLower(resp.Stderr + resp.Stdout)
	if strings.Contains(lower, "cannot cascade merge-back non-interactively") {
		t.Fatalf("old cascade pre-flight must not fire; stderr=%q", resp.Stderr)
	}
	assertContains(t, resp.Stderr, "blocks wrk --done")

	assertFileNotExists(t, req.ExternalWtDir)
	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertContains(t, depLog, "dep fix on external worktree")
	assertFileExists(t, req.WtDir)
}
```
