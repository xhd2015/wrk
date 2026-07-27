## Expected

- Non-zero exit (hard abort; replaces `warning: skipping nested main repo`).
- Stderr contains structured **`Error:`** prefix.
- Error body mentions nested main (path or basename) and cascade/nested-main context.
- Stderr must **not** treat this as a soft skip only: if `warning: skipping nested main`
  appears without a hard `Error:`, fail.
- External dep worktree still present; consumer linked worktree still present.
- Nested main path still present (not deleted).
- Pipe/default: no ANSI required for structure (assert no CSI on stderr).

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on nested main hard error; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if req.SecondRepo == "" {
		t.Fatal("SecondRepo (nested main path) must be set")
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "Error:") {
		t.Fatalf("nested main must hard-fail with structured %q; stderr:\n%s\nstdout:\n%s",
			"Error:", stderr, resp.Stdout)
	}
	// Prefer product wording; tolerate path-bearing cascade nested-main framing.
	lower := strings.ToLower(stderr)
	if !strings.Contains(lower, "nested") || !strings.Contains(lower, "main") {
		t.Fatalf("Error: body should mention nested main; stderr:\n%s", stderr)
	}
	nestedBase := filepath.Base(req.SecondRepo)
	if !strings.Contains(stderr, req.SecondRepo) && !strings.Contains(stderr, nestedBase) {
		// Also allow vendor/nested-main relative fragment.
		if !strings.Contains(stderr, "nested-main") && !strings.Contains(stderr, "vendor") {
			t.Fatalf("Error: must include nested main path %q or basename %q; stderr:\n%s",
				req.SecondRepo, nestedBase, stderr)
		}
	}
	// Soft-skip-only is the old behavior and must not satisfy this leaf.
	if strings.Contains(stderr, "warning: skipping nested main") &&
		!strings.Contains(stderr, "Error:") {
		t.Fatalf("old warn+skip nested main is insufficient; want hard Error:; stderr:\n%s", stderr)
	}

	if strings.Contains(stderr, "\x1b[") || strings.Contains(stderr, "\033[") {
		t.Fatalf("default pipe stderr must not require ANSI; got:\n%s", stderr)
	}

	assertCascadePreflightNoRemovals(t, req)
	assertFileExists(t, req.SecondRepo)
	// Cascade must not have removed external via warn+skip path.
	if req.ExternalWtDir != "" {
		assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)
	}
}
```
