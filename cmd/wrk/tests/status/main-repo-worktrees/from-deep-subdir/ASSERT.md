## Expected

- Exit code 0; stderr empty.
- Main block: absolute Dir (`statusNormalizePath(main)`); still has `Remote:`.
- Primary out-of-tree linked: Dir via `statusDirLine` (absolute for this depth).
- No `---- external ----` header.
- Rel would be `../../../..` (4 leading `..`) which exceeds the soft cap of 2.

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

	mainDir := statusDirLine(t, req.RepoDir, req.MainRepo)
	wantAbs := statusNormalizePath(t, req.MainRepo)
	if mainDir != wantAbs {
		t.Fatalf("fixture expectation: main Dir want absolute %q, got %q", wantAbs, mainDir)
	}

	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		},
		nil,
	))
}
```
