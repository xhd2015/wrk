---
label: e2e
explanation: product binary (shell launch; Capture cannot fake SHELL)
---

## Expected

- Exit code 0.
- Stdout is new worktree path (set-task contract unchanged).
- Old path gone; new path exists.
- Stderr contains install hint `wrk --bash-integration --install`.
- Fake interactive shell launched with cwd = newPath.
- Follow-up file empty (channel closed).

## Exit Code

- 0

```go
import (
	"regexp"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
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
	assertInstallHint(t, resp.Stderr)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, wantPath)
	assertFollowupEmpty(t, resp)
}
```
