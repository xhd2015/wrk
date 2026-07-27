---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Stdout is the default worktree path + `\n`.
- Worktree exists under `{WRK_HOME}/worktrees/`.
- Stderr empty preferred.

## Side Effects

- Native create; `--no-config` does not block create when paired with `--new`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantDashboardCreateWorktree(req)
	assertCreateOKPath(t, req, resp, err, wt)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
