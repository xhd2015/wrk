## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-feature-foo-2026-06-30` (slash → `-` in path token).
- Git branch name is `feature-foo-2026-06-30` (**no** `/`; slash sanitized for branch too).
- Branch `feature/foo-2026-06-30` does **not** exist.
- Worktree directory exists with linked `.git` file.
- Invariant: `filepath.Base(path) == "myrepo-" + branch`.

## Exit Code

- 0

```go
import "path/filepath"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	token := sanitizeBranchToken("feature/foo")
	wantBranch := branchName(token, wrkDate, 0)
	wantPath := worktreePath(req.WrkHome, "myrepo", token, wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.RepoDir, wantBranch)
	assertBranchNotExists(t, req.RepoDir, "feature/foo-"+wrkDate)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)
	assertWorktreeListContains(t, req.RepoDir, wantPath)

	// Path base must equal basename + "-" + branch (wrk-managed invariant).
	if filepath.Base(wantPath) != "myrepo-"+wantBranch {
		t.Fatalf("invariant broken: base(%q) != myrepo-%s", wantPath, wantBranch)
	}
}
```
