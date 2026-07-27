## Expected

- Non-zero exit code.
- Stderr contains `--dry-run is only valid with` (host-list validation).
- Stderr lists `--propagate-tags` among valid dry-run hosts.
- Stdout empty.

## Errors

- Bare `--dry-run` without a host mode is invalid; after P3 the host list
  includes `--propagate-tags` alongside existing hosts.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assertContains(t, resp.Stderr, "--dry-run is only valid with")
	assertContains(t, resp.Stderr, "--propagate-tags")
}
```
