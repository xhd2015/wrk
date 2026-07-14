## Expected

- Exit code 0.
- Appended external block has `Status: dirty (0 added, 1 changed, 0 renamed, 0 deleted)`.
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

	dirtyLine := statusLineForRepo(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockPlain(t, req.MainRepo, ".", "clean", "", true),
		appendedHealthyBlockPlain(t, req.MainRepo, req.WtDir, req.WtBranch, dirtyLine),
	))
}
```