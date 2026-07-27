## Expected

- `err` is nil.
- Dead/prunable path is in `Primary` (ListLinked membership), after main,
  before the subsequent linked entry.
- `External` is empty (dead path must not spill into external).

## Side Effects

- None (pure helper). Path existence is irrelevant for P1 membership.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertPartition(t, req, resp, err)
}
```
