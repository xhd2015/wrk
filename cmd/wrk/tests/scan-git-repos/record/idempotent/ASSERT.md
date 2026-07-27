## Expected

- Exit code 0.
- Stdout contains the resolved main repo absolute path exactly once (always-print of valid finds; not empty).
- `projects.json` still has exactly one entry for the main repo with `source: "scan"`.

## Side Effects

- No duplicate project entries after the second scan.
- Already-known does not suppress listing the live main path.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolveScanPath(t, req.MainRepo)
	n := countScanStdoutPathLines(t, resp.Stdout, want)
	if n != 1 {
		t.Fatalf("second scan must print known main path exactly once; count=%d stdout=%q want=%q", n, resp.Stdout, want)
	}
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
}
```
