## Expected Output

```
==== dep-update ====

  checkout  .
    module  example.com/app
      pin  example.com/lib  v1.0.0 -> v1.2.3
      skip tidy  (vendor/)

dep-update: updated 1, already 0, skipped 0 in 1 checkouts
```

## Expected

- Exit 0.
- Pin line for `example.com/lib` v1.0.0 -> v1.2.3.
- `skip tidy  (vendor/)` under the module.
- Summary `dep-update: updated 1, already 0, skipped 0 in 1 checkouts`.
- No `go mod tidy ok`; no `go mod vendor`.
- Require bumped; no go.sum; vendor/ has no `modules.txt`.
- Owner go.mod unchanged.

## Side Effects

- Same `--all` pin selection; tidy helper skips when vendor/ is present.

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
	assertNotContains(t, resp.Stdout, "would:")
	assertNoArgvDepHeader(t, resp.Stdout)
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update ====

  checkout  \.
    module  example\.com/app
      pin  example\.com/lib  v1\.0\.0 -> v1\.2\.3
      skip tidy  \(vendor/\)

dep-update: updated 1, already 0, skipped 0 in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "go mod tidy ok")
	assertNotContains(t, resp.Stdout, "go mod vendor")
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertNoTidyArtifacts(t, req)
	assertVendorUntouched(t, req.VendorDir)
	assertOwnerGoModUnchanged(t, req)
}
```
