## Expected

- Exit code 0.
- Stdout is **set-config dispatcher** usage mentioning `--set-config`, `--create`, `--show`, and help.
- Stdout points users toward create-level help (e.g. `--create --help`).
- Stdout is **not** the dedicated create UX dump (few or no create UX flags).
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- No `config.json` write required (help short-circuit).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSetConfigDispatcherHelp(t, resp.Stdout, resp.Stderr, resp.ExitCode)
}
```
