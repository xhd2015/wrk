## Expected

- Exit code 0.
- Stdout is the resolved main repo absolute path.
- `projects.json` is empty / has zero entries (print-only).
- Functional discovery works with `--no-cache` (does not require asserting cache files absent).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutExactPath(t, resp.Stdout, resolveScanPath(t, req.MainRepo))
	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)
}
```
