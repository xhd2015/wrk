## Expected

- Exit 0.
- Stdout is the default worktree path + `\n`.
- Worktree exists with linked `.git`.
- Does **not** require `--new` in argv.

## Side Effects

- Same as historical `wrk <dir>` create.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantDashboardCreateWorktree(req)
	assertCreateOKPath(t, req, resp, err, wt)
	assertBranchExists(t, req.MainRepo, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.MainRepo, wt)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
