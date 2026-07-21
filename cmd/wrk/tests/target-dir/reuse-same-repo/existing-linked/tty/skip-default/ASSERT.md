---
label: tty
explanation: requires `script` fake TTY; platform-specific
---

## Expected

- Exit code 0.
- Stdout (trimmed) equals the existing linked worktree path (lex-smallest / only one).
- No new worktree under `{WorkRoot}/target/`.
- Existing path still present and listed on source main.
- Combined output does **not** require skip prompt tokens (default auto-skip is silent).

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

	wantPath := req.WtDir
	// script(1) may multiplex TTY bytes; require path line present, prefer exact when clean.
	got := strings.TrimSpace(resp.Stdout)
	if got != wantPath && !strings.Contains(resp.Stdout, wantPath) {
		t.Fatalf("stdout should be/include existing path %q; stdout=%q stderr=%q", wantPath, resp.Stdout, resp.Stderr)
	}
	// Prefer exact path-only stdout when harness delivers clean separation.
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
	assertNotContains(t, combined, "Proceed?")
}
```
