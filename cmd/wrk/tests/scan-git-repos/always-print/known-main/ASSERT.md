## Expected

- Exit code 0.
- Stdout includes the resolved main absolute path exactly once (always-print; not silent for known).
- `projects.json` still has exactly one entry with `source: "scan"`.

## Side Effects

- Known path does not re-record or duplicate in projects.json.
- Listing is independent of the record gate.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolveScanPath(t, req.MainRepo)
	n := countScanStdoutPathLines(t, resp.Stdout, want)
	if n != 1 {
		t.Fatalf("known main must appear exactly once on stdout; count=%d stdout=%q want=%q", n, resp.Stdout, want)
	}
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
}
```
