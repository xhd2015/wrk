## Expected

- Exit code 0 (cascade removes `deps/foo`, then consumer merge-back succeeds).
- Manual linked worktree at `deps/foo` no longer exists (cascade via `scan_repo.Scan`).
- Consumer linked worktree no longer exists.
- Stdout contains `merged branch`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 (cascade removes deps/foo then consumer merge-back), got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	assertFileNotExists(t, req.DepsLinkedWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.DepsLinkedWtDir)
	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertContains(t, resp.Stdout, "merged branch")
}
```