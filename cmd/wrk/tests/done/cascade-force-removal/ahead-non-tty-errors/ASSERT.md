## Expected

- Exit code 0 (cascade auto-yes on non-TTY; no confirmation hard-guard).
- Dep fix merged into dep main.
- External dependency worktree under `external/` no longer exists.
- Consumer linked worktree no longer exists.
- Combined output has no `Proceed?`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 (cascade auto-yes on non-TTY), got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	combined := resp.Stdout + resp.Stderr
	assertNotContains(t, combined, "Proceed?")

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertContains(t, depLog, "dep fix on external worktree")

	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.ExternalWtDir)

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```
