## Expected Output

```
# module example.com/cli-multi-root (.)
would: go install ./cmd/rootbin
# module example.com/cli-multi-tools (tools)
would: go install ./cmd/toolbin
would: reinstall 2 binaries (0 skipped) across 2 modules
```

## Expected

- Exit code 0.
- Stdout is exactly the multi dry-run plan above (trailing newline on last line).
- Module order: scan-root module first (lex ModuleRoot), then nested `tools/`.
- Headers use full go.mod module path and RelDir relative to scan root
  (`.` for root, `tools` for nested).
- Summary includes `across 2 modules` (multi-only suffix; K>1).
- Stub binaries under GOBIN remain unchanged (dry-run).

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "" +
		"# module example.com/cli-multi-root (.)\n" +
		"would: go install ./cmd/rootbin\n" +
		"# module example.com/cli-multi-tools (tools)\n" +
		"would: go install ./cmd/toolbin\n" +
		"would: reinstall 2 binaries (0 skipped) across 2 modules\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertStubBinUnchanged(t, req.BinDir, "rootbin")
	assertStubBinUnchanged(t, req.BinDir, "toolbin")
}
```
