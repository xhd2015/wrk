## Expected

- Exit code 0.
- Stdout is the new worktree absolute path (trailing `\n`).
- Old operated worktree path gone; new path exists.
- Sibling shell cwd still exists.
- Follow-up file is empty (cwd gate: shell cwd still present).

## Exit Code

- 0

```go
import (
	"regexp"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slugify("new task"), 0)
	assert.Output(t, resp.Stdout, "---\nversion: 3\n---\n"+regexp.QuoteMeta(wantPath)+"\n")
	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertFileExists(t, req.RepoDir)
	assertFollowupEmpty(t, resp)
}
```
