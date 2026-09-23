
## Expected

- Non-zero exit; stdout empty.
- Stderr mentions mutual exclusion (`mutually exclusive`).
- No install runs.

## Errors

- `--install` is an exclusive mode; `--list` is not a partner.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertMutualExclusion(t, resp)
	assertContains(t, resp.Stderr, "--install")
	assertBinNotExists(t, req.BinDir, "tool")
}
```
