## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-{short-hash}-2026-06-30` where `{short-hash}` is the 7-char prefix recorded in Setup.
- Branch `{short-hash}-2026-06-30` exists and is checked out in the new worktree (not literal `HEAD`).
- Worktree directory exists with linked `.git` file.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if req.HashToken == "" {
		t.Fatal("req.HashToken must be set in Setup")
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", req.HashToken, wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.RepoDir, branchName(req.HashToken, wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName(req.HashToken, wrkDate, 0))
	assertWorktreeListContains(t, req.RepoDir, wantPath)
}
```