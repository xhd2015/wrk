
## Expected

- Non-zero exit.
- Stdout empty.
- Stderr matches the library wording: `--install` **requires a value**
  (same Varargs `WithMinimum(1)` family as `--bring`).

## Errors

- `--install` is a Varargs flag requiring at least one non-flag token per
  occurrence; a following flag is not a value.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "--install")
	assertContains(t, resp.Stderr, "requires a value")
	assertBinNotExists(t, req.BinDir, "tool")
}
```
