---
label: tty
explanation: requires `script` fake TTY; platform-specific
---

## Expected

- Exit code 0.
- Stdout refers to the lex-smallest existing path (`…/myrepo-main-{date}`, not `…-1`).
- No new worktree under `{WorkRoot}/target/`.
- Both prior worktrees remain.
- Prompt tokens not required under default auto-skip.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	smallest := req.WtDir
	other := req.ExternalWtDir2
	if smallest > other {
		// Defensive: identity is lex order of abs paths.
		smallest, other = other, smallest
	}

	got := strings.TrimSpace(resp.Stdout)
	if got != smallest && !strings.Contains(resp.Stdout, smallest) {
		t.Fatalf("stdout should be/include lex-smallest %q; stdout=%q stderr=%q", smallest, resp.Stdout, resp.Stderr)
	}
	// Must not report the larger path as the sole/primary stdout path when clean.
	if got == other {
		t.Fatalf("stdout must not be the non-smallest path %q", other)
	}

	assertFileExists(t, smallest)
	assertFileExists(t, other)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate))
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1"))
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-2"))
}
```
