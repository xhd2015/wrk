## Expected

- Exit code 0.
- Stdout: new worktree path, then same path from `pwd`.
- Old worktree path gone; new path exists with renamed branch.
- Stderr empty (confirm env bypasses TTY prompt text).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	slug := slugify("newslug")
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 0)
	assertPathThenChildStdout(t, resp.Stdout, wantPath, wantPath)

	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	oldBranch := branchNameWithTask("main", wrkDate, slugify("original"), 0)
	newBranch := branchNameWithTask("main", wrkDate, slug, 0)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchExists(t, req.MainRepo, newBranch)
	assertWorktreeListContains(t, req.MainRepo, wantPath)
}
```
