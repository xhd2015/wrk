
## Expected Output

```
# module example.com/cli-install-main-root (.)
would: go install ./cmd/mainbin
# module example.com/cli-install-main-tools (tools)
would: go install ./cmd/toolbin
would: install 2 binaries across 2 modules
```

## Expected

- Exit code 0 (compose accepted; **not** a mutual-exclusion error).
- Stdout is exactly the **main** multi-module install plan above.
- Plan does not mention `wtbin` or the worktree-only module path.
- No `skip:` line and no GOBIN stub anywhere: names were forced, not gated.

## Side Effects

- Dry-run only; no nested interactive shell, no GOBIN write.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertNotContains(t, resp.Stderr, "mutually exclusive")
	want := "" +
		"# module example.com/cli-install-main-root (.)\n" +
		"would: go install ./cmd/mainbin\n" +
		"# module example.com/cli-install-main-tools (tools)\n" +
		"would: go install ./cmd/toolbin\n" +
		"would: install 2 binaries across 2 modules\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "wtbin")
	assertNotContains(t, resp.Stdout, "cli-install-wt-root")
	assertNotContains(t, resp.Stdout, "skip:")
	assertBinNotExists(t, req.BinDir, "mainbin")
	assertBinNotExists(t, req.BinDir, "toolbin")
}
```
