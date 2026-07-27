## Expected

- `resp.IsWip` is false.
- `err` is nil.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.IsWip {
		t.Fatalf("IsWipSubject(%q) = true, want false", req.Subject)
	}
}
```
