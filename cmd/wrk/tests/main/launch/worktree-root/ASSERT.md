## Expected

- Exit code 0.
- Minimal launch UX (empty stdout; no install hint).
- Fake shell launched with cwd = **main repo root**, not the linked worktree path.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertMinimalLaunchUX(t, resp)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, req.MainRepo)
	// Shell must not land on the linked worktree path.
	if resolvePath(t, req.WtDir) == resolvePath(t, req.MainRepo) {
		t.Fatal("fixture error: linked worktree path equals main repo")
	}
}
```
