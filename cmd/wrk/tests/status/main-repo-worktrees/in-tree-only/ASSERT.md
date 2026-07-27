## Expected

- Exit code 0.
- Two primary blocks: main + in-tree linked (Dirs via statusDirLine from main root).
- No `---- external ----` header.
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

	master := masterField(t, req.MainRepo, "main", req.InTreeWtBranch)
	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			scanStatusBlockFromCwd(t, req.RepoDir, req.InTreeWtDir, "clean", master, false),
		},
		nil,
	))

	assertStdoutHasNoAppendedAbsDir(t, resp.Stdout, resolvePath(t, req.InTreeWtDir))
}
```
