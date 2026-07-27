## Expected

- Exit code 0.
- Primary blocks: main (Dir `.` + Remote) then out-of-tree wrk linked
  (`statusDirLine` Dir, typically `../.wrk/worktrees/…`, + Master identical).
- No `---- external ----` header (main-owned linked is primary, not external section).
- Blank line between blocks.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 2)
	assertNoExternalSectionHeader(t, resp.Stdout)

	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		},
		nil,
	))
}
```
