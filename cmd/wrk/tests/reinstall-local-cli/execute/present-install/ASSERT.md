
## Expected Output

```
go install ./cmd/tool
reinstalled 1, skipped 0, failed 0
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above (trailing newline on last line).
- No `would:` dry-run vocabulary on stdout.
- `$GOBIN/tool` is no longer the stub contents; it is executable and prints `tool-v1`.

## Side Effects

- Real `go install ./cmd/tool` with `Dir=moduleRoot` and `GOBIN=BinDir` replaces the stub.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "go install ./cmd/tool\nreinstalled 1, skipped 0, failed 0\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "would:")
	assertBinNotStub(t, req.BinDir, "tool")
	assertBinExecutable(t, req.BinDir, "tool")
	assertBinRuns(t, req.BinDir, "tool", "tool-v1")
}
```
