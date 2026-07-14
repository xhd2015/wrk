## Expected

- Exit code 0 on the second `wrk` invocation.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-main-2026-06-30-1`.
- Worktree directory `myrepo-main-2026-06-30-1` exists with linked `.git` file.
- Branch `main-2026-06-30-1` exists and is checked out in the new worktree (always-new-branch; not a reuse of the first).
- First worktree still has its own branch `main-2026-06-30` checked out (both branches are distinct new refs).
- Both `myrepo-main-2026-06-30` and `myrepo-main-2026-06-30-1` appear in `git worktree list`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 1)
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	firstPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertFileExists(t, firstPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	firstBranch := branchName("main", wrkDate, 0)
	secondBranch := branchName("main", wrkDate, 1)
	assertBranchExists(t, req.RepoDir, firstBranch)
	assertBranchExists(t, req.RepoDir, secondBranch)
	assertBranchCheckedOutInWorktree(t, firstPath, firstBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, secondBranch)
	assertWorktreeListContains(t, req.RepoDir, firstPath)
	assertWorktreeListContains(t, req.RepoDir, wantPath)
}
```
