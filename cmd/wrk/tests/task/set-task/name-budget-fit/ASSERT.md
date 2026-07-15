## Expected

- Exit 0; worktree renamed.
- New path basename and branch each ≤ 255 bytes.
- `filepath.Base(newPath) == basename + "-" + branch`.
- Soft-cap 64 slug alone would overflow path base for this basename — fitted slug is shorter.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	basename := strings.Repeat("r", 180)
	wt := strings.TrimSpace(resp.Stdout)
	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	// Old path gone
	assertFileNotExists(t, req.WtDir)

	base := filepath.Base(wt)
	if len(base) > 255 {
		t.Fatalf("path basename len %d > 255: %q", len(base), base)
	}
	br := gitOutputIsolated(t, wt, "rev-parse", "--abbrev-ref", "HEAD")
	if len(br) > 255 {
		t.Fatalf("branch len %d > 255: %q", len(br), br)
	}
	if base != basename+"-"+br {
		t.Fatalf("invariant: base=%q want=%q", base, basename+"-"+br)
	}
	full := slugify(req.SetTaskDesc)
	if utf8.RuneCountInString(full) < 20 {
		t.Fatalf("expected long soft-cap slug, got %q", full)
	}
	unfitted := basename + "-main-" + wrkDate + "-" + full
	if len(unfitted) <= 255 {
		t.Fatalf("fixture should overflow without fit: len=%d", len(unfitted))
	}
	if strings.HasSuffix(base, "-"+full) || strings.HasSuffix(br, "-"+full) {
		t.Fatalf("set-task should fit-shorten slug; base=%q branch=%q full=%q", base, br, full)
	}
	assertBranchExists(t, req.MainRepo, br)
	assertWorktreeListContains(t, req.MainRepo, wt)
}
```
