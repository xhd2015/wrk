## Expected

- Exit code 0.
- Stdout is exactly the resolved main repo absolute path (single line + trailing `\n`).
- `projects.json` is empty / has zero entries after cold scan (print-only).

## Side Effects

- `{WRK_HOME}/projects.json` is not created/updated by scan.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolveScanPath(t, req.MainRepo)
	assertStdoutExactPath(t, resp.Stdout, want)
	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)
}
```
