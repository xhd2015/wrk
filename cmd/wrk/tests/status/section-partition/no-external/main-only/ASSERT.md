## Expected

- `err` is nil.
- `Primary` is exactly `[main]` (normalized).
- `External` is empty.

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
