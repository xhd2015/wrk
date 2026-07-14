## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-main-2026-06-30` exactly.
- Worktree directory exists with a `.git` regular file (linked worktree layout).
- Branch `main-2026-06-30` exists and is checked out in the new worktree.
- `git worktree list` from source repo includes the new path.

## Side Effects

- `{WRK_HOME}/worktrees/` directory is created if missing.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.RepoDir, branchName("main", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.RepoDir, wantPath)

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```