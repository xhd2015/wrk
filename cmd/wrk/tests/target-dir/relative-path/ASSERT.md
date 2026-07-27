## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WorkRoot}/wt` exactly — the relative `wt` resolved against the
  process (shell) cwd `{WorkRoot}`, NOT against the source repo `{WorkRoot}/myrepo`.
- Worktree directory exists at `{WorkRoot}/wt` with a `.git` regular file (linked worktree).
- Branch `main-2026-06-30` exists in the source repo and is checked out in the worktree.
- `git worktree list` from `myrepo` contains `{WorkRoot}/wt`.
- No worktree created at `{WorkRoot}/myrepo/wt` (would exist if resolved against `<dir>`).
- Stderr is empty.

## Side Effects

- `{WorkRoot}/wt` directory created via `git worktree add`.

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

	// must NOT have been resolved against the source repo dir
	assertFileNotExists(t, filepath.Join(req.TargetDir, "wt"))

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
