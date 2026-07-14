## Expected

- Non-zero exit code.
- Stderr contains `--fetch is only valid with --projects or --status`.
- Stdout is empty.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertFetchInvalidModeStderr(t, resp)
}
```