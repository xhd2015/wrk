## Expected

- Exit code 0.
- Primary broken block has red ANSI on `error: <git stderr>` value only.
- `Dir:` label stays uncolored; Dir value follows `statusDirLine`.
- No `---- external ----` header (plain header N/A; primary-only fixture).
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

	errStatus := appendedErrorStatusColored(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			appendedDirLine(t, req.RepoDir, req.WtDir) + "\nStatus:       " + errStatus,
		},
		nil,
	))
}
```
