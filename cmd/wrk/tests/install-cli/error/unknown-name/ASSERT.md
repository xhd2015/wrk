
## Expected

- Non-zero exit; stdout empty (plan never starts).
- Stderr carries the `wrk: --install:` prefix, the requested name `nope`, and
  the `no install candidate` reason (shared with the named reinstall lookup).
- Nothing is installed.

## Errors

- The name does not match any discovered `cmd/<name>` or
  `script/<name>/install` candidate under the scan root.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "wrk: --install:")
	assertContains(t, resp.Stderr, "no install candidate")
	assertContains(t, resp.Stderr, "nope")
	assertBinNotExists(t, req.BinDir, "nope")
}
```
