## Expected

- `err` is nil.
- Worktree directory still exists.
- Main `HEAD` SHA equals pre-run snapshot (no merge applied).
- Worktree branch name is unchanged.

## Side Effects

- None (DryRun plan only).

## Errors

- None.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = resp
	assertErrIsNil(t, err)
	assertDirExists(t, req.WtDir)
	after := revParseHEAD(t, req.MainRepo)
	if after != req.MainHEADBefore {
		t.Fatalf("main HEAD mutated under DryRun: before %s after %s", req.MainHEADBefore, after)
	}
	gotBranch := currentBranch(t, req.WtDir)
	if gotBranch != req.WtBranch {
		t.Fatalf("worktree branch changed under DryRun: got %q want %q", gotBranch, req.WtBranch)
	}
}
```
