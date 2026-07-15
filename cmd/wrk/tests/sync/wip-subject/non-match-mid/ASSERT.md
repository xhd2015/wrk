## Expected

- `resp.IsWip` is false (match is prefix-only after trim; mid-string `wip:` does not count).
- `err` is nil.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.IsWip {
		t.Fatalf("IsWipSubject(%q) = true, want false (not a prefix)", req.Subject)
	}
}
```
