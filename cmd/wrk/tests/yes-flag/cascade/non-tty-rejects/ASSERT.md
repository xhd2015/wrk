## Expected

- Non-zero exit code.
- External dependency worktree still exists with ahead commit.
- Consumer linked worktree still exists.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (cascade guard on non-TTY), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertFileExists(t, req.ExternalWtDir)
	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertNotContains(t, depLog, "dep fix on external worktree")
	assertFileExists(t, req.WtDir)
}
```
