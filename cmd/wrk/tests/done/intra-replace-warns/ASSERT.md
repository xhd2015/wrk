## Expected

- Exit code 0 (relative intra-repo replace is silent under the default lenient guard).
- Stderr does **not** contain `tolerated` or the replace issue line for `submod`.
- Merge-back proceeded: the consumer linked worktree is removed and no longer
  registered.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 (relative intra-repo silent + proceed), got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	if strings.Contains(resp.Stderr, "tolerated") {
		t.Fatalf("relative intra-repo must be silent, got tolerated warn: stderr=%q", resp.Stderr)
	}
	submodAbs := filepath.Join(req.WtDir, "submod")
	if strings.Contains(resp.Stderr, "=> "+submodAbs) {
		t.Fatalf("relative intra-repo must not print replace issue line, stderr=%q", resp.Stderr)
	}

	// Merge-back ran and removed the worktree.
	assertFileNotExists(t, req.WtDir)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```
