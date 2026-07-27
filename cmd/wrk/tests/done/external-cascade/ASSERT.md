## Expected

- Non-zero exit code (local filesystem replace remains after cascade).
- Stderr contains `go.mod: =>` and the external dep worktree abs path.
- Stderr mentions the block (`blocks wrk --done`).
- External dependency worktree under `external/` no longer exists (cascade merge-back).
- Consumer linked worktree still exists (parent `--done` did not complete).

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (local replace guard), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "go.mod: => ")
	assertContains(t, resp.Stderr, req.ExternalWtDir)
	assertContains(t, resp.Stderr, "blocks wrk --done")
	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.MainRepo, req.ExternalWtDir)
	assertFileExists(t, req.WtDir)
}
```
