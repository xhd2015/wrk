## Expected Output

```
would: go run ./script/foo/install
would: reinstall 1 binaries (0 skipped)
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above.
- Stdout must **not** contain `go install ./cmd/foo` (script wins).
- Stub binary under GOBIN remains unchanged.

## Side Effects

- Dry-run only: no rewrite of stub bins.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go run ./script/foo/install\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "go install ./cmd/foo")
	assertStubBinUnchanged(t, req.BinDir, "foo")
}
```
