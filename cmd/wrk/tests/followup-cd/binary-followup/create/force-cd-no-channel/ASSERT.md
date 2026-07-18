## Expected

- Exit code 0 (fake shell exits 0).
- Stdout is the new worktree absolute path (create contract unchanged; trailing `\n`).
- Stderr contains install hint `wrk --bash-integration --install`.
- Fake interactive shell launched with cwd = new worktree.
- Follow-up file stays empty (channel was not open).

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
	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assert.Output(t, resp.Stdout, "---\nversion: 3\n---\n"+regexp.QuoteMeta(wantPath)+"\n")
	assertInstallHint(t, resp.Stderr)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, wantPath)
	assertFollowupEmpty(t, resp)
	assertFileExists(t, wantPath)
}
```
