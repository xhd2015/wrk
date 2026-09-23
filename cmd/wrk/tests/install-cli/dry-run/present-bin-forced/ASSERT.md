
## Expected Output

```
would: go install ./cmd/tool
would: install 1 binaries
```

## Expected

- Exit code 0; stdout is exactly the two lines above.
- No `skip:` line: `--install` never reports a bin as absent-from-binDir, so the
  plan is identical whether or not `$GOBIN/tool` already exists.
- The stub binary is unchanged (no mutation in dry-run).

## Side Effects

- Dry-run plan only; stub under GOBIN is not rewritten.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/tool\nwould: install 1 binaries\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "skip:")
	assertStubBinUnchanged(t, req.BinDir, "tool")
}
```
