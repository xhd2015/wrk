
## Expected

- Exit code 0.
- Stdout contains `go install ./cmd/tool` and ends with `installed 1, failed 0`.
- No `skip:` line even though the bin already existed: `--install` installs the
  requested name unconditionally.
- `$GOBIN/tool` is no longer the stub and prints `tool-ok`.

## Side Effects

- Real `go install` replaces the stub under the isolated GOBIN.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertContains(t, resp.Stdout, "go install ./cmd/tool")
	assertInstallSummary(t, resp.Stdout, 1, 0)
	assertNotContains(t, resp.Stdout, "skip:")
	assertBinNotStub(t, req.BinDir, "tool")
	assertBinRuns(t, req.BinDir, "tool", "tool-ok")
}
```
