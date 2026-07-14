## Expected

- Non-zero exit code (`-y` has no effect on cascade ahead/diverged confirmation in non-TTY mode).
- External dependency worktree still exists with the ahead commit.
- Consumer linked worktree still exists.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (-y must not bypass cascade guard on non-TTY), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertFileExists(t, req.ExternalWtDir)
	assertWorktreeListContains(t, req.DepPath, req.ExternalWtDir)

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertNotContains(t, depLog, "dep fix on external worktree")

	assertFileExists(t, req.WtDir)
}
```
