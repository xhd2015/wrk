## Expected

- Exit code 0 (soft skip — not a hard failure).
- Stdout (trimmed) equals `{plain}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree owned by the **dep** main repo.
- Branch in the dep repo is `main-{WRK_DATE}`.
- **No** `.gitignore` under the plain cwd (ensureGitignoreExternal skipped).
- Stderr contains `SKIP local dep replacement`, `is not a git repository`, and the abs plain cwd path.
- Stderr does **not** use the "not a dependency" / "not a go module" soft-skip wordings for this path.

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

	assertContains(t, resp.Stderr, "SKIP local dep replacement")
	assertContains(t, resp.Stderr, "is not a git repository")
	assertContains(t, resp.Stderr, req.RepoDir)
	assertNotContains(t, resp.Stderr, "not a dependency of any consumer module")
	assertNotContains(t, resp.Stderr, "is not a go module")
}
```
