## Expected Output

```
would: go install ./cmd/present
would: reinstall 1 binaries (0 skipped)
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above (trailing newline on last line).
- Stderr empty (or no hard error).
- Stub binary under GOBIN remains `stub-binary\n` (no real install).

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/present\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertStubBinUnchanged(t, req.BinDir, "present")
}
```
