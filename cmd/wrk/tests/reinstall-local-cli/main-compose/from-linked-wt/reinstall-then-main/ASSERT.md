## Expected Output

```
# module example.com/cli-main-root (.)
would: go install ./cmd/mainbin
# module example.com/cli-main-tools (tools)
would: go install ./cmd/toolbin
would: reinstall 2 binaries (0 skipped) across 2 modules
```

## Expected

- Exit code 0.
- Stdout matches MC1 exactly (flag order free for `--main` / `--reinstall-local`).
- Plan is main modules only (not worktree `wtbin`).
- Stubs unchanged.

## Side Effects

- Dry-run only.

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
	assertStubBinUnchanged(t, req.BinDir, "mainbin")
	assertStubBinUnchanged(t, req.BinDir, "toolbin")
	assertStubBinUnchanged(t, req.BinDir, "wtbin")
}
```
