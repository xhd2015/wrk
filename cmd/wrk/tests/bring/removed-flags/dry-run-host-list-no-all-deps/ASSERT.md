## Expected

- Non-zero exit code.
- Stderr contains `--dry-run is only valid with`.
- Stderr does **not** list `--all-deps` as a host.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "--dry-run is only valid with")
	assertNotContains(t, resp.Stderr, "--all-deps")
}
```
