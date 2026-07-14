## Expected

- Non-zero exit code.
- Stderr indicates cascade cannot proceed non-interactively (ahead/diverged linked worktree needs confirmation).
- External dependency worktree under `external/` still exists.
- Dep fix commit is still only on the external worktree branch (not merged into dep main).
- Consumer linked worktree still exists.

## Side Effects

- No `git worktree remove --force` on the external worktree.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (no force-remove), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	combined := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(combined, "cascade") &&
		!strings.Contains(combined, "ahead") &&
		!strings.Contains(combined, "diverged") &&
		!strings.Contains(combined, "confirmation") &&
		!strings.Contains(combined, "non-interactive") &&
		!strings.Contains(combined, "stdin is not a terminal") {
		t.Fatalf("expected cascade non-TTY guard error, got stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}

	assertFileExists(t, req.ExternalWtDir)
	assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertNotContains(t, depLog, "dep fix on external worktree")

	extLog := gitOutputIsolated(t, req.ExternalWtDir, "log", "--oneline")
	assertContains(t, extLog, "dep fix on external worktree")

	assertFileExists(t, req.WtDir)
}
```
