## Expected

- Exit code 0.
- Stdout is **show-level** usage mentioning `--set-config` and `--show`, plus JSON/config theme.
- Stdout is **not** a pure dump of config JSON as the help body.
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- No `config.json` write (help only).

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
