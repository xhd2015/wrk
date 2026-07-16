## Expected

- Exit code 0.
- Help text (stdout and/or stderr) contains `--reinstall-local`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--reinstall-local") {
		t.Fatalf("help must mention --reinstall-local; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, help, `<contains>
--reinstall-local
</contains>`)
}
```
