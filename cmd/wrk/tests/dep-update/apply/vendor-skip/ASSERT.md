## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/consumer
      pin  example.com/dep  v0.0.1 -> v0.0.2
      skip tidy  (vendor/)

dep-update: updated 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Pin applied (replace dropped; require @ latest).
- `skip tidy  (vendor/)` under the module.
- No `go mod tidy ok`; no `go mod vendor`.
- No go.sum; vendor/ has no `modules.txt`.

## Side Effects

- Pin only; tidy and `go mod vendor` skipped.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/consumer
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      skip tidy  \(vendor/\)

dep-update: updated 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "go mod tidy ok")
	assertNotContains(t, resp.Stdout, "go mod vendor")
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertNoTidyArtifacts(t, req)
	assertVendorUntouched(t, req.VendorDir)
}
```
