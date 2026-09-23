
## Expected Output

```
would: go install ./cmd/wtbin
would: install 1 binaries
```

## Expected

- Exit code 0.
- Without `--main`, the plan uses the **linked worktree** checkout: K=1 single
  module format (no `# module` headers, no `across … modules` suffix).
- Stdout is exactly the wt-only plan above; it does not mention `mainbin`,
  `toolbin`, or main module paths.
- No GOBIN mutation.

## Side Effects

- Dry-run only.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/wtbin\nwould: install 1 binaries\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "mainbin")
	assertNotContains(t, resp.Stdout, "toolbin")
	assertNotContains(t, resp.Stdout, "cli-install-main-root")
	assertNotContains(t, resp.Stdout, "across")
	assertBinNotExists(t, req.BinDir, "wtbin")
}
```
