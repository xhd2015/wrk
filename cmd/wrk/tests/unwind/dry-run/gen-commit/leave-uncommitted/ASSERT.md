## Expected Output

```
==== unwind (dry-run) ====
would: peel .
  would: leave 1 file uncommitted (use --add-all if necessary)
  would: generate commit message and commit staged changes
```

## Expected

- Exit code **0** (leave-N is plan language, not an error).
- Peel display `.`.
- Stdout contains locked leave-N line for N=1 (singular `file`):
  `would: leave 1 file uncommitted (use --add-all if necessary)`.
- Leave line appears under the peel and before/with generate/commit plan language.
- Stdout does **not** contain `would: git add -A` (no `--add-all`).
- Zero mutations: HEAD unchanged; untracked `DIRTY` still present.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	wantLeave := leaveLine(req.LeaveN)
	if !strings.Contains(resp.Stdout, wantLeave) {
		t.Fatalf("missing leave-N line %q\nstdout:\n%s", wantLeave, resp.Stdout)
	}
	assertContainsInOrder(t, resp.Stdout,
		peelLine("."),
		wantLeave,
		"generate",
		"commit",
	)
	if strings.Contains(resp.Stdout, "would: git add -A") {
		t.Fatalf("without --add-all must not plan git add -A; stdout:\n%s", resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
	// Untracked dirt still present (zero mutation).
	assertFileExists(t, filepath.Join(req.RepoDir, "DIRTY"))
}
```
