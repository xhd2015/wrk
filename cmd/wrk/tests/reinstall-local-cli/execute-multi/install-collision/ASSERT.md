## Expected

- Non-zero exit code (cross-module install×install hard plan error before execute).
- Stderr mentions bin name `samebin`.
- Stderr identifies both claiming modules (path bases `mod-a` and `mod-b`).
- No successful execute summary (`reinstalled` not required on stdout; prefer
  empty or no plan success summary).
- No `go install` progress lines implying a completed multi execute plan.

## Errors

- Same bin cannot be Action=install from two modules under one plan — fails
  during planning, before any `go install` / `go run`.

## Side Effects

- Plan-only failure: stub bin must not be rewritten by go install.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertContains(t, resp.Stderr, "samebin")
	assertContains(t, resp.Stderr, "mod-a")
	assertContains(t, resp.Stderr, "mod-b")
	assertNotContains(t, resp.Stdout, "reinstalled")
	assertNotContains(t, resp.Stdout, "go install")
	assertStubBinUnchanged(t, req.BinDir, "samebin")
}
```
