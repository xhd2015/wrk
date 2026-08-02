## Expected

- `err` is nil.
- `resp.MainAbs` is the absolute main repository path.
- `resp.MainAbs` is **not** equal to the linked worktree path.
- `resp.MainAbs` equals the seeded MainRepo (cleaned / symlink-resolved).

## Side Effects

- None (read-only resolve).

## Errors

- None.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.MainAbs == "" {
		t.Fatal("MainAbs empty")
	}
	assertPathEqual(t, resp.MainAbs, req.MainRepo)
	assertPathNotEqual(t, resp.MainAbs, req.WtDir)
}
```
