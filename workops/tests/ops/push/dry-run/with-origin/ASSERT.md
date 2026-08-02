## Expected

- `err` is nil.
- Origin `refs/heads/main` SHA equals pre-run snapshot (no network mutation).
- Local main HEAD still present (no destructive local change required).

## Side Effects

- None (DryRun no-op for network).

## Errors

- None.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = resp
	assertErrIsNil(t, err)
	after := revParseRef(t, req.OriginBare, "refs/heads/main")
	if after != req.OriginHEADBefore {
		t.Fatalf("origin/main mutated under DryRun: before %s after %s", req.OriginHEADBefore, after)
	}
}
```
