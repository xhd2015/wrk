## Expected

- Exit code 0.
- Scan block for `Dir:          .` (no `Master:`).
- Appended full block for external wt with absolute `Dir` and `Master: identical`.
- Blank line between blocks.
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
		scanStatusBlockPlain(t, req.MainRepo, ".", "clean", "", true),
		appendedHealthyBlockPlain(t, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
	))
}
```