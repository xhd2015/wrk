## Expected

- `err` is nil.
- `resp.MainAbs` equals the main checkout path (cleaned / abs).

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
}
```
