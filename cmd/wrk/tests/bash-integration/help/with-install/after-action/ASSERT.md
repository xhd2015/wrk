## Expected

- Exit code 0.
- Stdout is bash-integration usage (not install status lines).
- Stdout is **not** the bash.sh script.
- Stderr is empty.
- No `bash.sh` and no profile markers written.

## Side Effects

- Help short-circuit; install must not run.

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
	assertNotContains(t, resp.Stdout, "bash integration:")
}
```
