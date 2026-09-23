
## Expected Output

```
would: go install ./cmd/tool
would: install 1 binaries
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above (summary last, trailing `\n`).
- The plan installs the name even though `$GOBIN/tool` does not exist: no
  `skip:` line anywhere, proving the binDir gate is off.
- `$GOBIN/tool` was not created (dry-run mutates nothing).

## Side Effects

- Dry-run plan only; no `go install` and no GOBIN write.

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
	assertNotContains(t, resp.Stdout, "reinstall")
	assertBinNotExists(t, req.BinDir, "tool")
}
```
