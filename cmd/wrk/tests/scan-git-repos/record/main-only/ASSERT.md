## Expected

- Exit code 0.
- Stdout is exactly the main repo absolute path (not the worktree path).
- `projects.json` has exactly one entry: the main path with `source: "scan"`.
- Linked worktree path is not recorded.

## Side Effects

- Only `RepoTypeMain` paths become project records.

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
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")
	assertScanProjectNotRecorded(t, req.WrkHome, req.WtDir)
}
```
