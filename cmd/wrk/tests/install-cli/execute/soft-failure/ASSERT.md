
## Expected

- Exit code **0** even though the install failed (soft-failure policy, same as
  `--reinstall-local`).
- Stdout contains `go install ./cmd/broken` and ends with
  `installed 0, failed 1`.
- Stderr contains the child compiler noise and the wrk-owned soft-failure line
  `warning: install finished with 1 failed`.
- No binary is left under GOBIN.

## Side Effects

- Failed `go install`; GOBIN unchanged.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertContains(t, resp.Stdout, "go install ./cmd/broken")
	assertInstallSummary(t, resp.Stdout, 0, 1)
	assertContains(t, resp.Stderr, "warning: install finished with 1 failed")
	assertBinNotExists(t, req.BinDir, "broken")
}
```
