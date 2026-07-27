## Expected

- Non-zero exit code.
- Stdout empty.
- Stderr contains `--dry-run is only valid with`.
- Stderr mentions `--reinstall-local` among valid dry-run hosts
  (exact host-list wording/order is implementer-owned; substring lock only).

## Errors

- Bare `--dry-run` without a host mode is invalid; P2 extends the host list to
  include `--reinstall-local`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "--dry-run is only valid with")
	assertContains(t, resp.Stderr, "--reinstall-local")
}
```
