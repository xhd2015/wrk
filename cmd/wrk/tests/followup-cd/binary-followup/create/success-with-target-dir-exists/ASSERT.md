## Expected

- Exit code 0.
- Stdout is the worktree absolute path
  `{WorkRoot}/target/myrepo-main-2026-06-30` (trailing `\n`).
- Worktree exists under the existing target parent with default naming.
- Follow-up file is empty even though shell cwd was FakeHome / user home
  (target-dir create never writes follow-up `cd`).
- Stderr empty (follow-ups are file-only on the binary surface).

## Side Effects

- `{WorkRoot}/target/myrepo-main-2026-06-30` created as a linked worktree.
- No `cd` line written to `WRK_FOLLOWUP_FILE`.
- Default `{WRK_HOME}/worktrees/...` is not used.

## Exit Code

- 0

```go
import (
	"regexp"
	"path/filepath"
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
	wantPath := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assert.Output(t, resp.Stdout, "---\nversion: 3\n---\n"+regexp.QuoteMeta(wantPath)+"\n")
	assertFollowupEmpty(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("binary stderr should be empty (follow-ups are file-only), got %q", resp.Stderr)
	}
	assertFileExists(t, wantPath)
	// Must not have fallen back to default WRK_HOME spawn location.
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
}
```
