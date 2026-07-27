## Expected

- `err` is nil.
- `Primary` is `[main, linkedLate, linkedEarly]` — **ListLinked order**, not
  path sort (`aaa-early` would sort before `zzz-late`).
- `External` is empty.

## Side Effects

- None (pure helper).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPartition(t, req, resp, err)
	// Extra pin: path sort of the two linked paths is Early then Late,
	// so a buggy path-sort primary would fail the WantPrimary check above.
	if len(req.WantPrimary) != 3 {
		t.Fatalf("fixture contract: want 3 primary paths, got %d", len(req.WantPrimary))
	}
}
```
