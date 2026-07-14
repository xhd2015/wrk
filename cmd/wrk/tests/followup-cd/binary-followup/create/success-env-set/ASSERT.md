## Expected

- Exit code 0.
- Stdout is the new worktree absolute path (trailing `\n`).
- Follow-up file contains exactly `cd <same-worktree-abs>` with trailing newline
  (home gate opens: shell cwd was FakeHome / user home).
- Stderr empty (no follow-up on stderr from binary).

## Exit Code

- 0

```go
import (
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
	assert.Output(t, resp.Stdout, "---\nversion: 2\n---\n"+wantPath+"\n")
	assertFollowupCD(t, resp, wantPath)
	if resp.Stderr != "" {
		t.Fatalf("binary stderr should be empty (follow-ups are file-only), got %q", resp.Stderr)
	}
	assertFileExists(t, wantPath)
}
```
