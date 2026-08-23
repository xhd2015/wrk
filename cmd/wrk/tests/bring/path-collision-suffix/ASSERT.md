## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-1`.
- Plain occupied `{consumerTop}/external/mydep` remains (no Policy A reuse).
- Branch in dep repo is `main-{WRK_DATE}-1` (joint path+branch `-N`).

## Exit Code

- 0

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	blockedPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertFileExists(t, blockedPath)
	assertFileExists(t, filepath.Join(blockedPath, "not-a-worktree"))

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 1)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)

	wantBranch := branchName("main", wrkDate, 1)
	assertBranchNotExists(t, req.DepPath, branchName("main", wrkDate, 0))
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)
	assertNotContains(t, resp.Stderr, "already exists under external/")
	assertNotContains(t, resp.Stderr, "reusing")
}
```
