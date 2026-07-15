## Expected

- Exit code 0.
- Scan block for main unchanged (Dir via statusDirLine).
- Appended minimal block: `statusDirLine` Dir (from porcelain path) + `Status: prunable`.
- No `Branch`/`Commit`/`Master:` on appended block.
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

	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
		appendedMinimalBlockPlain(t, req.RepoDir, req.WtDir, "prunable"),
	))
}
```
