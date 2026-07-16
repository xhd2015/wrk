## Expected Output

```
# module example.com/cli-main-root (.)
would: go install ./cmd/mainbin
# module example.com/cli-main-tools (tools)
would: go install ./cmd/toolbin
would: reinstall 2 binaries (0 skipped) across 2 modules
```

## Expected

- Exit code 0 (compose accepted; dry-run completed — **not** mutual exclusion).
- Stdout is exactly the **main** multi-module dry-run plan above.
- Plan does **not** mention `wtbin` or `cli-wt-root` (worktree-only module).
- Stub binaries under GOBIN remain unchanged (dry-run; no nested shell install path).

## Side Effects

- Dry-run only: no rewrite of stub bins. No nested interactive shell.

## Errors

- Must **not** fail with `mutually exclusive` (compose is the product under test).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertNotContains(t, resp.Stderr, "mutually exclusive")
	want := "" +
		"# module example.com/cli-main-root (.)\n" +
		"would: go install ./cmd/mainbin\n" +
		"# module example.com/cli-main-tools (tools)\n" +
		"would: go install ./cmd/toolbin\n" +
		"would: reinstall 2 binaries (0 skipped) across 2 modules\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "wtbin")
	assertNotContains(t, resp.Stdout, "cli-wt-root")
	assertStubBinUnchanged(t, req.BinDir, "mainbin")
	assertStubBinUnchanged(t, req.BinDir, "toolbin")
	assertStubBinUnchanged(t, req.BinDir, "wtbin")
}
```
