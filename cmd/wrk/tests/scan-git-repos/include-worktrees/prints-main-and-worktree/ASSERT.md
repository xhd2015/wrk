## Expected

- Exit code 0.
- Stdout contains the resolved main absolute path at least once.
- Stdout contains the resolved worktree absolute path at least once.
- Each path appears at most once (dedup).
- `projects.json` is empty / has zero entries (print-only).
- Worktree is printed only with --include-worktrees; never recorded.

## Side Effects

- Worktrees are list-only when `--include-worktrees` is set.
- Default without the flag is covered by `record/main-only` (worktree not printed).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	wantMain := resolveScanPath(t, req.MainRepo)
	wantWt := resolveScanPath(t, req.WtDir)
	nMain := countScanStdoutPathLines(t, resp.Stdout, wantMain)
	nWt := countScanStdoutPathLines(t, resp.Stdout, wantWt)
	if nMain != 1 {
		t.Fatalf("main path must appear exactly once; count=%d stdout=%q want=%q", nMain, resp.Stdout, wantMain)
	}
	if nWt != 1 {
		t.Fatalf("worktree path must appear exactly once with --include-worktrees; count=%d stdout=%q want=%q", nWt, resp.Stdout, wantWt)
	}
	// Print-only: scan never mutates projects.json (main or worktree).
	assertScanProjectsCount(t, req.WrkHome, 0)
}
```
