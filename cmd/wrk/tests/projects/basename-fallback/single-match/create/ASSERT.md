## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-main-2026-06-30`.
- Worktree directory exists and is a linked git worktree of the **saved** repo.
- Branch `main-2026-06-30` exists in the saved repo and is checked out in the new worktree.
- `git worktree list` from the saved repo includes the new path.

## Side Effects

- Worktree created under `{WRK_HOME}/worktrees/` from the saved project path, not from cwd.

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
	assertBranchExists(t, req.MainRepo, branchName("main", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.MainRepo, wantPath)

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```