
## Expected

- Non-zero exit; stdout empty.
- Stderr mentions `mutually exclusive` and names **both** flags (`--install`
  and `--reinstall-local`), so the user knows which pair conflicts.
- No install runs.

## Errors

- `--install` (forced, named) and `--reinstall-local` (binDir-gated) are
  distinct modes and cannot be combined.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "mutually exclusive")
	assertContains(t, resp.Stderr, "--install")
	assertContains(t, resp.Stderr, "--reinstall-local")
	assertBinNotExists(t, req.BinDir, "tool")
}
```
