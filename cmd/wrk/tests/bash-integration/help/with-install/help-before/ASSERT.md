## Expected

- Exit code 0.
- Stdout is bash-integration usage.
- No `bash.sh` and no profile markers written.
- Stderr is empty.

## Side Effects

- Help short-circuit regardless of flag order; install must not run.

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
