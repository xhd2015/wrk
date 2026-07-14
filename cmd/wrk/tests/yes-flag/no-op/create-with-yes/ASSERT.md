## Expected

- Exit code 0.
- Stdout is the new linked worktree absolute path.
- Worktree exists under `WRK_HOME` with expected branch name.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	req.WtDir = wantPath
	req.WtBranch = branchName("main", wrkDate, 0)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, wantPath)
}
```
