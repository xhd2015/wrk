## Expected

- Exit code 0.
- Stdout is **dedicated create** usage mentioning `--set-config`, `--create`, and multiple UX flags (`--new-window`, `--open-in-agent`, plus further UX flags).
- Stdout ends with trailing `\n`.
- Stderr is empty.
- Help short-circuits before "requires at least one create UX flag".

## Side Effects

- No `config.json` write (help only; no UX flags that would merge).

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
