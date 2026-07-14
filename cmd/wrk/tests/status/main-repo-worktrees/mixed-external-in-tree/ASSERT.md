## Expected

- Exit code 0.
- Scan blocks: `.` and `wt-linked` (relative, `Master:` on in-tree linked).
- One appended block: external wt with absolute `Dir` (not duplicated in scan).
- Total three blocks separated by blank lines.
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
	assertStdoutBlocksSeparated(t, resp.Stdout, 3)

	inTreeMaster := masterField(t, req.MainRepo, "main", req.InTreeWtBranch)
	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockPlain(t, req.MainRepo, ".", "clean", "", true),
		scanStatusBlockPlain(t, req.InTreeWtDir, req.InTreeWtRel, "clean", inTreeMaster, false),
		appendedHealthyBlockPlain(t, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
	))
}
```