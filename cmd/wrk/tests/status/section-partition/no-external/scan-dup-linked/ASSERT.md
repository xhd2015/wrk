## Expected

- `err` is nil.
- `Primary` is `[main, inTree]` — in-tree linked once only (no scan duplicate).
- `External` is empty (must not classify the linked path as external).

## Side Effects

- None (pure helper).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertPartition(t, req, resp, err)
}
```
