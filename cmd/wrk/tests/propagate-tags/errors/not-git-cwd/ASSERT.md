## Expected

- Non-zero exit code.
- Stderr mentions the cwd is not a git repository.

## Errors

- Must not plan source releases or would-update lines on success path.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertContains(t, resp.Stderr, "not a git repository")
}
```
