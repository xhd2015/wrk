
## Expected Output

```
would: go run ./script/tool/install
would: install 1 binaries
```

## Expected

- Exit code 0; stdout is exactly the two lines above.
- Resolution is the same as `--reinstall-local`: the `script/<name>/install`
  candidate wins and uses the `go-run-install` method (`go run`, not
  `go install`).
- No `skip:` line; no mutation under GOBIN.

## Side Effects

- Dry-run only.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := "would: go run ./script/tool/install\nwould: install 1 binaries\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "go install")
	assertBinNotExists(t, req.BinDir, "tool")
}
```
