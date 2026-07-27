## Expected

- Exit code 0.
- Stdout is **show-level** usage (same level as `--show --help`).
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
	assertSetConfigShowHelp(t, resp.Stdout, resp.Stderr, resp.ExitCode)
	assertSetConfigNoConfigWrite(t, req.WrkHome)
}
```
