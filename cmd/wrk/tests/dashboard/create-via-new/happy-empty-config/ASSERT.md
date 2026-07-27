---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Stdout is exactly the default worktree path + trailing `\n`.
- Worktree exists under `{WRK_HOME}/worktrees/` with linked `.git` file.
- Branch `main-2026-06-30` is checked out in the new worktree.
- `git worktree list` from the source repo includes the new path.
- Stderr empty.

## Side Effects

- Native create only (same side effects as historical bare `wrk`).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantDashboardCreateWorktree(req)
	assertCreateOKPath(t, req, resp, err, wt)
	assertBranchExists(t, req.RepoDir, branchName("main", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wt, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.RepoDir, wt)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
