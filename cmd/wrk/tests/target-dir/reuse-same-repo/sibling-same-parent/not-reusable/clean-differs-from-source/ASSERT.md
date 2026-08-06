## Expected

- Exit code 0.
- Stdout is new path `{WorkRoot}/target/myrepo-main-{date}` (not the sibling).
- New worktree at preferred name; sibling still present with differing HEAD.
- No Policy B banner (`would reuse`, `skip creating`, `already has a linked worktree`,
  `refusing non-interactive`).

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantNew := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assertStdoutExactPath(t, resp.Stdout, wantNew)
	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertBranchCheckedOutInWorktree(t, wantNew, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.TargetDir, wantNew)

	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)
	// Sibling still differs from source (MainRepo stashed source HEAD in SETUP).
	if req.MainRepo != "" {
		sibHEAD := revParseHEAD(t, req.WtDir)
		if sibHEAD == req.MainRepo {
			t.Fatalf("sibling HEAD should still differ from source %s", req.MainRepo)
		}
	}

	if strings.TrimSpace(resp.Stdout) == req.WtDir {
		t.Fatalf("stdout must not reuse clean-differs sibling %q", req.WtDir)
	}
	assertNoPolicyBBanner(t, resp.Stdout+resp.Stderr)
}
```
