## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-main-2026-06-30` exactly — same as `wrk` from `myrepo`.
- Worktree directory exists with a `.git` regular file (linked worktree layout).
- Branch `main-2026-06-30` exists and is checked out in the new worktree.
- `git worktree list` from source repo includes the new path.
- Side effects match `wrk` no-args create from `myrepo` (worktree path, branch, git worktree list).

## Side Effects

- `{WRK_HOME}/worktrees/` directory is created if missing.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.TargetDir, branchName("main", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```