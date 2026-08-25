## Expected

- Exit code 0.
- Stdout is **bash-integration** usage mentioning `--bash-integration`, `--install`, `--uninstall`, `--status`, and help.
- Stdout is **not** the bash.sh script dump.
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- No `bash.sh` or profile marker writes.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertBashIntegrationUsage(t, resp.Stdout, resp.Stderr, resp.ExitCode)
	assertBashIntegrationHelpNoFilesystemWrite(t, req)
}
```
