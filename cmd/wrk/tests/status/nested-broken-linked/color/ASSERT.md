## Expected

- Exit code 0.
- Gray ANSI header `---- external ----` between primary and external sections (P3).
- Broken external block has red ANSI on `error: <git stderr>` value only.
- `Dir:` label stays uncolored.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 4)

	errStatus := scanErrorStatusColored(t, req.WtDir)
	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternalColored(t,
		[]string{
			colorScanRootBlockPlain(t, req.MainRepo),
		},
		[]string{
			colorScanStatusBlockPlain(t, req.DepPath, "tools/good"),
			colorScanStatusBlockPlain(t, req.ConsumerTop, "vendor/host"),
			scanBrokenBlockPlain("vendor/host/broken-wt", errStatus),
		},
	))
}
```