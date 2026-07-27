## Expected

- Exit code 0.
- Stdout is skill-level usage mentioning `--list`, `--show`, and `--install`.
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSkillUsageStdout(t, resp.Stdout, resp.Stderr, resp.ExitCode)
}
```
