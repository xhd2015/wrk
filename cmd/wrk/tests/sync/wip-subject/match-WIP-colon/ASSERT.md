## Expected

- `resp.IsWip` is true.
- `err` is nil.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if !resp.IsWip {
		t.Fatalf("IsWipSubject(%q) = false, want true", req.Subject)
	}
}
```
