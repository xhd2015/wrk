## Expected

- Exit code 0.
- Two scan blocks: main + in-tree linked (Dirs via statusDirLine from main root).
- No appended external section.
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

	master := masterField(t, req.MainRepo, "main", req.InTreeWtBranch)
	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
		scanStatusBlockFromCwd(t, req.RepoDir, req.InTreeWtDir, "clean", master, false),
	))

	assertStdoutHasNoAppendedAbsDir(t, resp.Stdout, resolvePath(t, req.InTreeWtDir))
}
```
