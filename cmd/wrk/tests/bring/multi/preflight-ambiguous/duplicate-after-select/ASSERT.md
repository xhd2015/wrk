## Expected

- Non-zero exit code.
- Stderr: **at most one** `Select [` for the ambiguous `mydep` basename (resolve-once).
- Stderr indicates duplicate bring path (prefer `wrk: duplicate --bring path: <abs>` with the selected abs).
- **No** `will bring:` plan (error during preflight before plan).
- Prefer empty stdout; no `external/` directory created.

## Side Effects

- Fail before any worktree create when two args resolve to the same absolute path after selection.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for duplicate after select, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	// One Select max for the single ambiguous basename.
	if n := strings.Count(resp.Stderr, "Select ["); n > 1 {
		t.Fatalf("expected at most one Select prompt, got %d stderr=%q", n, resp.Stderr)
	}
	// Prefer exactly one listing for mydep when prompt path is taken.
	if n := strings.Count(resp.Stderr, `Multiple projects match "mydep":`); n > 1 {
		t.Fatalf("expected at most one mydep listing, got %d stderr=%q", n, resp.Stderr)
	}

	// Duplicate class (stable preferred wording).
	se := strings.ToLower(resp.Stderr)
	ok := strings.Contains(se, "duplicate") ||
		strings.Contains(se, "same") ||
		strings.Contains(se, "already") ||
		strings.Contains(se, "twice") ||
		strings.Contains(se, "repeated")
	if !ok {
		t.Fatalf("stderr should indicate duplicate bring path; got %q", resp.Stderr)
	}

	// Prefer error names the resolved abs path.
	dupAbs := multiPreflightAbs(t, req.SelectedSavedRepo)
	if !strings.Contains(resp.Stderr, dupAbs) && !strings.Contains(resp.Stderr, req.SelectedSavedRepo) {
		// Soft: wording may omit path but still mention duplicate; log path for implementer.
		t.Logf("note: preferred error includes resolved abs %s; stderr=%q", dupAbs, resp.Stderr)
	}

	// No plan on preflight hard error.
	assertNotContains(t, resp.Stderr, "will bring:")

	// Prefer fail before materializing.
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))

	// Prefer empty stdout (no external path lines).
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("expected empty stdout on duplicate preflight error, got %q", resp.Stdout)
	}
}
```
