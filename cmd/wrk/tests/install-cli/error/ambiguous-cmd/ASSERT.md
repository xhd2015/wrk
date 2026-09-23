
## Expected

- Non-zero exit; stdout empty.
- Stderr carries the `wrk: --install:` prefix, the bin name `dup`, the word
  `ambiguous`, and both candidate paths sorted lexicographically
  (`./cmd/dup`, `./cmd/nested/dup`).
- The ambiguity is reported instead of picking one candidate silently
  (discovery omits the ambiguous bin, and named lookup surfaces it).

## Errors

- Same bin basename discovered twice under `cmd/`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "wrk: --install:")
	assertContains(t, resp.Stderr, "ambiguous")
	assertContains(t, resp.Stderr, "./cmd/dup")
	assertContains(t, resp.Stderr, "./cmd/nested/dup")
}
```
