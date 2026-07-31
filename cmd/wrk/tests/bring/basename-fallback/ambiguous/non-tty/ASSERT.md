## Expected

- Non-zero exit code.
- Stderr mentions multiple matches for `mydep` **exactly once** (no double dump from duplicate preflight + apply resolve).
- Stderr lists both candidate absolute paths (lexicographically sorted).
- Stdout is empty.
- No external worktree created under either saved dep repo.

## Errors

- Ambiguous basename without TTY cannot proceed.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}

	// Listing once: preflight/apply must not print the candidate block twice.
	if n := strings.Count(resp.Stderr, `Multiple projects match "mydep":`); n != 1 {
		t.Fatalf("expected exactly one match listing, got %d stderr=%q", n, resp.Stderr)
	}

	sorted := sortedSavedPaths(t, req.DepPath, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "mydep":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
</contains>`
	assert.Output(t, resp.Stderr, tmpl)

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertFileNotExists(t, wantPath)
	assertWorktreeListNotContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.SecondRepo, wantPath)
}
```
