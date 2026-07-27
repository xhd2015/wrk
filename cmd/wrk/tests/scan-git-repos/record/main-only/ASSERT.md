## Expected

- Exit code 0.
- Stdout is exactly the main repo absolute path (not the worktree path).
- `projects.json` is empty / has zero entries (print-only).
- Linked worktree is not printed by default (main-only leaf).

## Side Effects

- Only `RepoTypeMain` paths are printed by default.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	wantMain := resolveScanPath(t, req.MainRepo)
	assertStdoutExactPath(t, resp.Stdout, wantMain)
	assertNotContains(t, resp.Stdout, resolveScanPath(t, req.WtDir))
	// Print-only: scan never mutates projects.json (main or worktree).
	assertScanProjectsCount(t, req.WrkHome, 0)
}
```
