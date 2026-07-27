---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Stdout is worktree path with task slug after the date.
- Worktree exists under `{WRK_HOME}/worktrees/`.
- Does **not** require `--new`.

## Side Effects

- Create with `-t` / `--task` as today.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantDashboardCreateWorktreeWithTask(req, "ship feature")
	assertCreateOKPath(t, req, resp, err, wt)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
