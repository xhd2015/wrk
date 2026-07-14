## Expected

- Exit code 0.
- Stdout is **set-config dispatcher** usage (same level as `--help`).
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertSetConfigDispatcherHelp(t, resp.Stdout, resp.Stderr, resp.ExitCode)
}
```
