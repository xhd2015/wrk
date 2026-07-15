## Expected

- Exit code 0.
- Stdout is empty (no newly recorded paths).
- `projects.json` still has exactly one entry for the main repo with `source: "scan"`.

## Side Effects

- No duplicate project entries after the second scan.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != "" {
		t.Fatalf("second scan should print no newly recorded paths, got stdout=%q", resp.Stdout)
	}
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
}
```
