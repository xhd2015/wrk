## Expected Output

```
skip: missing (not in <BinDir>)
would: reinstall 0 binaries (1 skipped)
```

## Expected

- Exit code 0 (successful plan with N=0 installs).
- Stdout: one `skip:` line for `missing` naming resolved GOBIN path, then summary
  `would: reinstall 0 binaries (1 skipped)\n`.
- No `would: go install` / `would: go run` lines.
- Optional `warning:` on stderr when nothing to reinstall is allowed (not required).

## Side Effects

- None (no bin stubs; dry-run).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := fmt.Sprintf("skip: missing (not in %s)\nwould: reinstall 0 binaries (1 skipped)\n", req.BinDir)
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "would: go install")
	assertNotContains(t, resp.Stdout, "would: go run")
}
```
