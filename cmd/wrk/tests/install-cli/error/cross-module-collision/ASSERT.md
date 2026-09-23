
## Expected

- Non-zero exit; stdout empty (no install attempted).
- Stderr carries the `wrk: --install:` prefix, the bin name `same`, the phrase
  `multiple modules`, and both claiming module paths (`mod-a`, `mod-b`).
- Unlike the bare `--reinstall-local` plan, the collision is detected at named
  lookup even though neither bin exists in GOBIN.

## Errors

- One bin name is claimed by two modules in the same multi-module plan.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "wrk: --install:")
	assertContains(t, resp.Stderr, "same")
	assertContains(t, resp.Stderr, "multiple modules")
	assertContains(t, resp.Stderr, "mod-a")
	assertContains(t, resp.Stderr, "mod-b")
}
```
