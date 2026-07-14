## Expected

- Non-zero exit code (`--confirm-from-stdin` does not apply to cascaded ahead/diverged worktrees on non-TTY).
- External dependency worktree still exists.
- Dep fix commit was not merged into dep main.
- Consumer linked worktree still exists.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (option A cascade guard), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertFileExists(t, req.ExternalWtDir)
	assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertNotContains(t, depLog, "dep fix on external worktree")

	extLog := gitOutputIsolated(t, req.ExternalWtDir, "log", "--oneline")
	assertContains(t, extLog, "dep fix on external worktree")

	assertFileExists(t, req.WtDir)
}
```
