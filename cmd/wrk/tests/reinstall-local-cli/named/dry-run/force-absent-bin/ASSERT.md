## Expected Output

```
would: go install ./cmd/tool
would: reinstall 1 binaries (0 skipped)
```

## Expected

- Exit code 0.
- Stdout is exactly the two lines above.
- No `skip:` line (named mode does not apply the binDir gate).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/tool\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "skip:")
}
```
