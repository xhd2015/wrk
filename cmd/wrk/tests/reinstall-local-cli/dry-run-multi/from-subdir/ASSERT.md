
## Expected Output

```
# module example.com/cli-scan-root (.)
would: go install ./cmd/rootbin
# module example.com/cli-scan-tools (tools)
would: go install ./cmd/toolbin
would: reinstall 2 binaries (0 skipped) across 2 modules
```

## Expected

- Exit code 0.
- Cwd is under `pkg/sub` (no go.mod); plan still covers **both** modules under
  the git toplevel (not walk-up failure / not single nested module only).
- Stdout is exactly the multi dry-run plan above.
- Stub binaries under GOBIN remain unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins. Git fixture stays under WorkRoot.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "" +
		"# module example.com/cli-scan-root (.)\n" +
		"would: go install ./cmd/rootbin\n" +
		"# module example.com/cli-scan-tools (tools)\n" +
		"would: go install ./cmd/toolbin\n" +
		"would: reinstall 2 binaries (0 skipped) across 2 modules\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertStubBinUnchanged(t, req.BinDir, "rootbin")
	assertStubBinUnchanged(t, req.BinDir, "toolbin")
}
```
