## Expected

- Exit code 0.
- Appended broken block has red ANSI on `error: <git stderr>` value only.
- `Dir:` label stays uncolored; Dir value follows `statusDirLine`.
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

	errStatus := appendedErrorStatusColored(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
		appendedDirLine(t, req.RepoDir, req.WtDir)+"\nStatus:       "+errStatus,
	))
}
```
