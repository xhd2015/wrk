## Expected Output

```
would: go install ./cmd/wtbin
would: reinstall 1 binaries (0 skipped)
```

## Expected

- Exit code 0.
- Without `--main`, plan uses the **linked worktree** checkout (K=1 single-mod
  format: no `# module` headers; summary without `across … modules`).
- Stdout is exactly the wt-only plan above (`wtbin`).
- Plan does **not** mention `mainbin`, `toolbin`, or main module paths
  (proves contrast with compose leaves).
- Stubs unchanged.

## Side Effects

- Dry-run only. Sealed single-mod dry-run format preserved for K=1.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/wtbin\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "mainbin")
	assertNotContains(t, resp.Stdout, "toolbin")
	assertNotContains(t, resp.Stdout, "cli-main-root")
	assertNotContains(t, resp.Stdout, "across")
	assertStubBinUnchanged(t, req.BinDir, "mainbin")
	assertStubBinUnchanged(t, req.BinDir, "toolbin")
	assertStubBinUnchanged(t, req.BinDir, "wtbin")
}
```
