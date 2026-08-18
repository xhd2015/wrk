## Expected Output

```
==== dep-update (dry-run) ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/consumer
      would: pin  example.com/dep  v0.0.1 -> v0.0.2
      would: skip tidy  (vendor/)

dep-update: would update 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Dry-run banner; `would: pin` and `would: skip tidy  (vendor/)`.
- No `would: go mod tidy`; no bare apply pin lines.
- go.mod unchanged; no go.sum; vendor/ untouched.

## Side Effects

- None: dry-run must not mutate go.mod or vendor/.

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
==== dep-update \(dry-run\) ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/consumer
      would: pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      would: skip tidy  \(vendor/\)

dep-update: would update 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "would: go mod tidy")
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
	assertVendorUntouched(t, req.VendorDir)
}
```
