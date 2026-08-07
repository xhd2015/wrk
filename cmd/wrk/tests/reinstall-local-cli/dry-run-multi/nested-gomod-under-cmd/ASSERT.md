
## Expected Output

```
# module example.com/cli-nested-cmd-root (.)
would: go install ./foo
# module example.com/cli-nested-cmd-mod (cmd)
would: reinstall 1 binaries (0 skipped) across 2 modules
```

## Expected

- Exit code 0.
- Stdout is exactly the multi dry-run plan above (trailing newline on last line).
- **Post-re-root** install path: `would: go install ./foo` (nearest go.mod is
  the nested `cmd/` module; rebased RelPath is `./foo`).
- Must **not** print unre-rooted `would: go install ./cmd/foo` alone as the
  install line (that is the pre-fix discovery RelPath under the parent plan
  module).
- Module `# module` headers may still attribute the item to the **plan**
  (parent) module; this assert only locks path strings + multi format, not
  plan-time ownership rewrite.
- Nested `cmd/` module block may be empty (current discovery only looks under
  that module's own `./cmd/...` / `./script/...`).
- Summary: N=1 install, M=0 skipped, across 2 modules (K=2).
- Stub binary under GOBIN remains unchanged (dry-run).

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "" +
		"# module example.com/cli-nested-cmd-root (.)\n" +
		"would: go install ./foo\n" +
		"# module example.com/cli-nested-cmd-mod (cmd)\n" +
		"would: reinstall 1 binaries (0 skipped) across 2 modules\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	// Explicit anti-regression: unre-rooted parent RelPath must not be the install line.
	assertNotContains(t, resp.Stdout, "would: go install ./cmd/foo")
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
