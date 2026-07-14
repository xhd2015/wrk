## Expected

- Exit code 0.
- Stdout is **dedicated create** usage (help short-circuits before merge).
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- `config.json` remains absent — help must not merge even when a UX flag is co-present.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertSetConfigCreateHelp(t, resp.Stdout, resp.Stderr, resp.ExitCode)
	assertSetConfigNoConfigWrite(t, req.WrkHome)
}
```
