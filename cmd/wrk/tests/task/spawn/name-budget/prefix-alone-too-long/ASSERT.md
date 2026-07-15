## Expected

- Exit non-zero.
- Stderr mentions name / path / too long / budget (flexible tokens).
- No worktree under WRK_HOME for this basename (no silent chop of basename to force fit).

## Exit Code

- non-zero

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when prefix alone exceeds budget; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Intentional name-budget UX (not only raw OS ENAMETOOLONG / "File name too long").
	if !strings.Contains(resp.Stderr, "wrk:") {
		t.Fatalf("stderr should be a wrk error (wrk: …), got %q", resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	intentional := strings.Contains(low, "255") ||
		strings.Contains(low, "budget") ||
		strings.Contains(low, "component") ||
		strings.Contains(low, "name too long") ||
		(strings.Contains(low, "too long") && (strings.Contains(low, "path") || strings.Contains(low, "branch") || strings.Contains(low, "worktree")))
	if !intentional {
		t.Fatalf("stderr should explain intentional name-budget failure (255/budget/component/…), got %q", resp.Stderr)
	}
	basename := longBasename(overBudgetBasenameLen)
	prefix40 := basename
	if len(prefix40) > 40 {
		prefix40 = prefix40[:40]
	}
	// Must not create a chopped-basename worktree.
	entries, _ := filepath.Glob(filepath.Join(req.WrkHome, "worktrees", "*"))
	for _, e := range entries {
		b := filepath.Base(e)
		if strings.HasPrefix(b, prefix40) {
			t.Fatalf("must not create worktree by chopping basename; found %q", e)
		}
	}
	out := strings.TrimSpace(resp.Stdout)
	if out != "" {
		if _, stErr := os.Stat(out); stErr == nil {
			t.Fatalf("must not leave a created path on budget error: %q", out)
		}
	}
}
```

