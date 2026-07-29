## Expected

- Exit code 0.
- Stdout (trimmed) equals `{plain}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree owned by the **dep** main repo.
- Branch in the dep repo is `main-{WRK_DATE}`.
- **No** `.gitignore` under the plain cwd.
- Stderr does **not** contain `SKIP local dep replacement` (analyse skipped by `--no-dep`).
- Stderr does **not** contain `mod tidy`.

## Exit Code

- 0

```go
import (
	"path/filepath"
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

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	assertFileNotExists(t, filepath.Join(req.ConsumerTop, ".gitignore"))

	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
	assertNotContains(t, resp.Stderr, "mod tidy")
}
```
