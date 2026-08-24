## Expected

- Non-zero exit.
- Stderr mentions `nope` and that there is no install candidate.
- Stdout empty (or no would: plan).

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertContains(t, resp.Stderr, "nope")
	assertContains(t, resp.Stderr, "no install candidate")
}
```
