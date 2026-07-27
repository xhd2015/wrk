## Expected

- `err` is nil.
- `ScanRoot` is the **main repo** absolute path (`…/mainrepo`), not the linked
  worktree path (`…/linked-wt`).
- Multi plan modules live under mainrepo (root + tools); both bins install.

## Side Effects

- Fixture git worktree under WorkRoot only.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertScanPlanOK(t, req, resp, err)
	// Extra lock: ScanRoot must not be the linked worktree WorkDir.
	if resolvePath(t, resp.ScanRoot) == resolvePath(t, req.WorkDir) {
		t.Fatalf("ScanRoot must be main repo, not linked worktree WorkDir %q", req.WorkDir)
	}
}
```
