
## Expected Output

```
go install ./cmd/rootbin
go install ./cmd/toolbin
reinstalled 2, skipped 0, failed 0
```

## Expected

- Exit code 0.
- Stdout is exactly the three lines above (trailing newline on last line).
- Module order: scan-root module first, then nested `tools/` (progress for
  `rootbin` before `toolbin`).
- No `# module` headers and no `would:` / `across` dry-run vocabulary (execute
  uses the same progress/summary format as single-mod).
- `$GOBIN/rootbin` and `$GOBIN/toolbin` are real executables (not stubs) and
  print `rootbin-ok` / `toolbin-ok` respectively — prove both modules ran with
  `Dir` set per module.

## Side Effects

- Real `go install` for each module's candidate with `Dir=ModuleRoot` and
  `GOBIN=BinDir` replaces both stubs.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "" +
		"go install ./cmd/rootbin\n" +
		"go install ./cmd/toolbin\n" +
		"reinstalled 2, skipped 0, failed 0\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "would:")
	assertNotContains(t, resp.Stdout, "# module")
	assertNotContains(t, resp.Stdout, "across")
	assertBinNotStub(t, req.BinDir, "rootbin")
	assertBinNotStub(t, req.BinDir, "toolbin")
	assertBinExecutable(t, req.BinDir, "rootbin")
	assertBinExecutable(t, req.BinDir, "toolbin")
	assertBinRuns(t, req.BinDir, "rootbin", "rootbin-ok")
	assertBinRuns(t, req.BinDir, "toolbin", "toolbin-ok")
}
```
