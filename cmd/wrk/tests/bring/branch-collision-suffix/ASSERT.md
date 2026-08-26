## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep` (path not bumped).
- Branch in dep repo is `main-{WRK_DATE}-1` (preferred `main-{date}` was taken).
- Prefixed legacy-style branch `mydep-main-{date}` / `mydep-main-{date}-1` must not be required.
- Stderr contains `warning: branch main-{WRK_DATE} exists; using main-{WRK_DATE}-1`.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)

	wantBranch := branchName("main", wrkDate, 1)
	assertBranchExists(t, req.DepPath, branchName("main", wrkDate, 0))
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchNotExists(t, req.DepPath, "mydep-"+branchName("main", wrkDate, 1))
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	assertContains(t, resp.Stderr, "warning: branch "+branchName("main", wrkDate, 0)+" exists; using "+wantBranch)
}
```
