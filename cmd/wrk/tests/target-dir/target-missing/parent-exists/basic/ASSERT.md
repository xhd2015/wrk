## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WorkRoot}/wt` exactly — the worktree is spawned at the fixed
  `<target-dir>` path with no default-naming suffix on the path.
- Worktree directory exists at `{WorkRoot}/wt` with a `.git` regular file (linked worktree).
- Branch `main-2026-06-30` exists in the source repo and is checked out in the worktree (new `-b`).
- `git worktree list` from `myrepo` contains `{WorkRoot}/wt`.
- The worktree does NOT live under `{WRK_HOME}/worktrees`.
- Stderr is empty.

## Side Effects

- `{WorkRoot}/wt` directory created (git worktree add).
- No directory created under `{WRK_HOME}/worktrees`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := filepath.Join(req.WorkRoot, "wt")
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	branch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.TargetDir, branch)
	assertBranchCheckedOutInWorktree(t, wantPath, branch)
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	// worktree must NOT live under WRK_HOME
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
