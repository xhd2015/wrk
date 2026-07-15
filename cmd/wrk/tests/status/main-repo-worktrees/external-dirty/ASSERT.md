## Expected

- Exit code 0.
- Primary out-of-tree block has `Status: dirty (0 added, 1 changed, 0 renamed, 0 deleted)`.
- Primary `Dir` follows `statusDirLine(main, wt)` (relative when ≤2 leading `..`).
- No `---- external ----` header.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 2)
	assertNoExternalSectionHeader(t, resp.Stdout)

	dirtyLine := statusLineForRepo(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, dirtyLine),
		},
		nil,
	))
}
```
