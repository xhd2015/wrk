---
label: e2e
explanation: product binary CLI integration (process boundary)
---


## Expected

- Exit code 0.
- Stdout contains the wrk-owned progress line `go install ./cmd/tool` and ends
  with the execute summary `installed 1, failed 0`.
- `$GOBIN/tool` exists, is executable, and prints `tool-ok`.
- No `skip:` line (the name was forced, not gated on preinstalled state).

## Side Effects

- Real `go install` writes `$GOBIN/tool` (isolated per leaf).

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
	assertBinExecutable(t, req.BinDir, "tool")
	assertBinRuns(t, req.BinDir, "tool", "tool-ok")
}
```
