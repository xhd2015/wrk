## Expected

- Exit code 0.
- Stdout is **dedicated create** usage (same level as `--create --help`).
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- No `config.json` write.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertSetConfigCreateHelp(t, resp.Stdout, resp.Stderr, resp.ExitCode)
	assertSetConfigNoConfigWrite(t, req.WrkHome)
}
```
