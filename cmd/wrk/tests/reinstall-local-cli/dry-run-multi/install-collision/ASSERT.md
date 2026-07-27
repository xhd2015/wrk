
## Expected

- Non-zero exit code (cross-module install×install hard error).
- Stderr mentions bin name `samebin`.
- Stderr identifies both claiming modules (path bases `mod-a` and `mod-b`
  and/or module basenames `cli-coll-a` / `cli-coll-b` — require path bases).
- No successful multi dry-run summary (`across` / `would: reinstall` not
  required on stdout; prefer empty or no plan success summary).

## Errors

- Same bin cannot be Action=install from two modules under one plan.

## Side Effects

- Dry-run planning only; stub bin may remain but must not be rewritten by go install.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertContains(t, resp.Stderr, "samebin")
	assertContains(t, resp.Stderr, "mod-a")
	assertContains(t, resp.Stderr, "mod-b")
	assertNotContains(t, resp.Stdout, "across")
	assertStubBinUnchanged(t, req.BinDir, "samebin")
}
```
