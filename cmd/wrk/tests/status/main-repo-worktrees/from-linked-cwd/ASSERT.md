## Expected

- Exit code 0.
- Single scan block for `Dir:          .` from the linked worktree cwd (with `Master:`).
- No appended section (no second block with absolute external path).
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
	assertStdoutBlocksSeparated(t, resp.Stdout, 1)

	master := masterField(t, req.MainRepo, "main", req.WtBranch)
	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockPlain(t, req.WtDir, ".", "clean", master, false),
	))

	assertStdoutHasNoAppendedAbsDir(t, resp.Stdout, resolvePath(t, req.WtDir))
}
```