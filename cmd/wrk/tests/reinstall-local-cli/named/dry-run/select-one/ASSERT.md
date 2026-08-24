## Expected Output

```
would: go install ./cmd/keep
would: reinstall 1 binaries (0 skipped)
```

## Expected

- Exit 0.
- Only `keep` appears; `drop` is absent from stdout.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go install ./cmd/keep\nwould: reinstall 1 binaries (0 skipped)\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "drop")
}
```
