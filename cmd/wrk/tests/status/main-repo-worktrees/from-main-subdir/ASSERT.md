## Expected

- Exit code 0; stderr empty.
- Main block: `Dir:          ../..` (not `.`); still has `Remote:`.
- Primary out-of-tree linked: Dir via `statusDirLine(subdir, wt)` (often absolute — three `..`).
- No `---- external ----` header.
- Branch/Commit/Status/Master content otherwise same as main-root status.

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
	if mainDir != "../.." {
		t.Fatalf("fixture expectation: main Dir want ../.., got %q (cwd=%s main=%s)",
			mainDir, req.RepoDir, req.MainRepo)
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
