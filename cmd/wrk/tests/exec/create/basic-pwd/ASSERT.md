## Expected

- Exit code 0.
- Stdout is exactly two lines: new worktree absolute path, then the same path from `pwd`.
- Worktree exists as a linked git worktree; branch `main-2026-06-30` checked out.
- Stderr empty.

## Side Effects

- `{WRK_HOME}/worktrees/myrepo-main-2026-06-30` created.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertPathThenChildStdout(t, resp.Stdout, wantPath, wantPath)

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
