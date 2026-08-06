---
label: e2e, tty
explanation: requires `script` fake TTY for Policy B multi skip prompt; platform-specific
---

## Expected

- Exit code 0.
- Stdout refers to the lex-smallest reusable path (`worktree-A`, not `worktree-B`).
- No new worktree under `{WorkRoot}/target/myrepo-main-{date}`.
- Both prior siblings remain.
- Combined output: `wrk: warning:`, `would reuse`, `skip creating`, primary path;
  multi: `also present` (or equivalent) listing the other reusable path.

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

	smallest := req.WtDir
	other := req.ExternalWtDir2
	if smallest > other {
		smallest, other = other, smallest
	}

	got := strings.TrimSpace(resp.Stdout)
	if got != smallest && !strings.Contains(resp.Stdout, smallest) {
		t.Fatalf("stdout should be/include lex-smallest %q; stdout=%q stderr=%q", smallest, resp.Stdout, resp.Stderr)
	}
	if got == other {
		t.Fatalf("stdout must not be the non-smallest path %q", other)
	}

	assertFileExists(t, smallest)
	assertFileExists(t, other)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate))
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1"))

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "wrk: warning:")
	assertContains(t, combined, "would reuse")
	assertContains(t, combined, "skip creating")
	assertContains(t, combined, smallest)
	// Multi awareness: also present wording and/or other path.
	if !strings.Contains(combined, other) {
		t.Fatalf("expected other reusable path in multi output; combined=%q", combined)
	}
	if !strings.Contains(combined, "also present") && !strings.Contains(combined, "also-present") && !strings.Contains(combined, "also") {
		t.Fatalf("expected multi also-present style wording; combined=%q", combined)
	}
}
```
