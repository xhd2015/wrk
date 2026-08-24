## Expected

- Non-zero exit.
- Stderr mentions that names are only valid on the exclusive `--reinstall-local` command.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertContains(t, resp.Stderr, "names are only valid")
	assertContains(t, resp.Stderr, "exclusive")
}
```
