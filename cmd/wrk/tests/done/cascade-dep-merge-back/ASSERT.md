## Expected

- Exit code 0 (cascade -y cascade yes merges ahead dep; consumer finishes).
- Dep fix committed on the external worktree is merged into dep main.
- External dependency worktree under `external/` no longer exists.
- Consumer linked worktree no longer exists.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 (cascade dep merge-back -y cascade yes), got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertContains(t, depLog, "dep fix on external worktree")

	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.ExternalWtDir)

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```
