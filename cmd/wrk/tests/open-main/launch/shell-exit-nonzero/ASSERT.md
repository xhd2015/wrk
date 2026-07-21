## Expected

- Exit code 42 (propagated from fake shell via ExitCodeError pattern like `--cd`).
- Minimal launch UX still applies (empty stdout; no install hint).
- Fake shell launched in main repo root.

## Exit Code

- 42

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 42 {
		t.Fatalf("expected exit 42 from shell, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertMinimalLaunchUX(t, resp)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, req.MainRepo)
}
```
