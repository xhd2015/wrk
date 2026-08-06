---
label: e2e, tty
explanation: requires `script` fake TTY for Policy B skip prompt; platform-specific
---

## Expected

- Exit code 0.
- Stdout (trimmed) equals the reusable sibling path (only one).
- No new worktree under `{WorkRoot}/target/` named `myrepo-main-{date}`.
- Sibling still present and listed on source main.
- Combined stdout+stderr contains Policy B tokens: `wrk: warning:`, `would reuse`,
  sibling path, and `skip creating` (skip-creating style prompt).
- Do not require ANSI color sequences.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := req.WtDir
	got := strings.TrimSpace(resp.Stdout)
	if got != wantPath && !strings.Contains(resp.Stdout, wantPath) {
		t.Fatalf("stdout should be/include reusable path %q; stdout=%q stderr=%q", wantPath, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "\n") || strings.TrimSpace(resp.Stdout) == wantPath {
		if strings.TrimSpace(resp.Stdout) == wantPath {
			assertStdoutExactPath(t, resp.Stdout, wantPath)
		}
	}

	assertFileExists(t, wantPath)
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	newUnderTarget := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assertFileNotExists(t, newUnderTarget)

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "wrk: warning:")
	assertContains(t, combined, "would reuse")
	assertContains(t, combined, "skip creating")
	assertContains(t, combined, wantPath)
}
```
