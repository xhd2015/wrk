## Expected

- `resp.IsWip` is false.
- `err` is nil.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.IsWip {
		t.Fatalf("IsWipSubject(%q) = true, want false", req.Subject)
	}
}
```
