## Expected

- Exit code 0.
- The main absolute path appears on stdout exactly once (no double-print from serve+refresh).
- `projects.json` is empty / has zero entries (print-only).

## Side Effects

- In-run dedup: same abs path printed at most once.

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
		t.Fatalf("main path must appear exactly once; count=%d stdout=%q want=%q", n, resp.Stdout, want)
	}
	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)
}
```
