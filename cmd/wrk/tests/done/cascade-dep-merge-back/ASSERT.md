## Expected

- Non-zero exit (local replace guard after cascade — same as `external-cascade`).
- Dep fix **was** merged into dep main (cascade auto-yes ran; pre-flight guard no longer blocks).
- External dependency worktree removed.
- Consumer linked worktree still exists.
- Stderr mentions replace blocking `--done`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (replace guard after cascade), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "blocks wrk --done")
	assertContains(t, resp.Stderr, "go.mod: => ")

	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertContains(t, depLog, "dep fix on external worktree")

	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.ExternalWtDir)

	assertFileExists(t, req.WtDir)
}
```
