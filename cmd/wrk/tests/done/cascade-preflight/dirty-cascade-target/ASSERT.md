## Expected

- Non-zero exit.
- Stderr includes **`Error:`** and dirty/uncommitted (or clean) language.
- Mentions the external worktree path or basename (cascade target context).
- **No removals**: external + consumer still present.
- Prefer all-or-nothing preflight (D2): nothing removed even if phase headers print.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on dirty cascade target; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if req.ExternalWtDir == "" {
		t.Fatal("ExternalWtDir must be set")
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "Error:") {
		// Some dirty paths may still surface without Error: today — require it for D2 framing.
		t.Fatalf("dirty cascade preflight must include structured %q; stderr:\n%s\nstdout:\n%s",
			"Error:", stderr, resp.Stdout)
	}
	combinedLower := strings.ToLower(stderr + "\n" + resp.Stdout)
	if !strings.Contains(combinedLower, "uncommitted") && !strings.Contains(combinedLower, "clean") &&
		!strings.Contains(combinedLower, "dirty") {
		t.Fatalf("expected dirty/uncommitted language; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	extBase := filepath.Base(req.ExternalWtDir)
	if !strings.Contains(stderr, req.ExternalWtDir) && !strings.Contains(stderr, extBase) {
		// Path may appear only in combined streams; prefer stderr.
		if !strings.Contains(resp.Stdout, req.ExternalWtDir) && !strings.Contains(resp.Stdout, extBase) {
			t.Fatalf("error must mention external path %q or base %q; stderr:\n%s\nstdout:\n%s",
				req.ExternalWtDir, extBase, stderr, resp.Stdout)
		}
	}

	assertCascadePreflightNoRemovals(t, req)
	assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)
}
```
