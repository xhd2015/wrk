## Expected

- Exit code 0.
- Stdout is **create-level** help (not merely dispatcher), because `--create` is present with help.
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- No `config.json` write.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSetConfigCreateHelp(t, resp.Stdout, resp.Stderr, resp.ExitCode)
	assertSetConfigNoConfigWrite(t, req.WrkHome)
}
```
