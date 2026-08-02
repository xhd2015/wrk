## Expected

- `err` is nil.
- Worktree directory still exists.
- Main `HEAD` SHA equals pre-run snapshot (no merge applied).
- Worktree branch name is unchanged.
- Sync=true under DryRun does not mutate either (same spirit as Sync=false dry-run).

## Side Effects

- None (DryRun plan only; Sync composition must not apply under DryRun).

## Errors

- None.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = resp
	if !req.Sync {
		t.Fatal("fixture must set Sync=true for this leaf")
	}
	if !req.DryRun {
		t.Fatal("fixture must set DryRun=true for this leaf")
	}
	assertErrIsNil(t, err)
	assertDirExists(t, req.WtDir)
	after := revParseHEAD(t, req.MainRepo)
	if after != req.MainHEADBefore {
		t.Fatalf("main HEAD mutated under DryRun+Sync: before %s after %s", req.MainHEADBefore, after)
	}
	gotBranch := currentBranch(t, req.WtDir)
	if gotBranch != req.WtBranch {
		t.Fatalf("worktree branch changed under DryRun+Sync: got %q want %q", gotBranch, req.WtBranch)
	}
}
```
