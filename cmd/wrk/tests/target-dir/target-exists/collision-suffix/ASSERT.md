## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WorkRoot}/target/myrepo-main-2026-06-30-1` exactly.
- Worktree directory exists at that path with a `.git` regular file (linked worktree).
- Branch `main-2026-06-30-1` exists in the source repo and is checked out in the worktree.
- `git worktree list` from `myrepo` contains the `-1` path.
- The pre-existing colliding sub-dir `myrepo-main-2026-06-30` (suffix 0) is left untouched.
- Stderr is empty.

## Side Effects

- `{WorkRoot}/target/myrepo-main-2026-06-30-1` created via `git worktree add`.
- No directory created under `{WRK_HOME}/worktrees`.

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1")
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	branch := branchName("main", wrkDate, 1)
	assertBranchExists(t, req.TargetDir, branch)
	assertBranchCheckedOutInWorktree(t, wantPath, branch)
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	// suffix-0 name was left as the pre-created empty dir (collision avoided by -1)
	assertFileExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate))

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
