
## Expected Output

```
would: go install ./cmd/tool
would: install 1 binaries
```

## Expected

- Exit code 0; stdout is exactly the two lines above.
- `other` is not planned and not mentioned anywhere in stdout.
- Exactly one install action is counted (the unrequested candidate is dropped,
  not installed and not skipped).

## Side Effects

- Dry-run only; no GOBIN write.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/tool\nwould: install 1 binaries\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "other")
	assertBinNotExists(t, req.BinDir, "tool")
}
```
