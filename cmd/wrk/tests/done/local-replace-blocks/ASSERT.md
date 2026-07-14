## Expected

- Non-zero exit code (extra-repo filesystem replace blocks `--done`).
- Stderr contains `go.mod: => <wtDir>/external/foo`.
- Stderr mentions the block (`blocks wrk --done`).
- Consumer linked worktree still exists (merge-back did not run).

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "go.mod: => "+filepath.Join(req.WtDir, "external", "foo"))
	assertContains(t, resp.Stderr, "blocks wrk --done")
	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```
