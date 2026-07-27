## Expected

- `err` is nil.
- `Primary` is `[main]`.
- `External` is the three nested paths in **lexicographic** normalized-path
  order: `external/child`, `task-hub`, `tools/child` — not the scan input order.

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
