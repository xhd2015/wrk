## Expected

- Exit code 0 (broken worktree does not abort the run).
- Primary main block unchanged (Dir via statusDirLine).
- Primary minimal block: `statusDirLine` Dir + `Status: error: <git stderr>` only.
- No `Branch`/`Commit`/`Master:` on broken block.
- No `---- external ----` header.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 2)
	assertNoExternalSectionHeader(t, resp.Stdout)

	errLine := appendedErrorStatusPlain(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			appendedMinimalBlockPlain(t, req.RepoDir, req.WtDir, errLine),
		},
		nil,
	))
}
```
