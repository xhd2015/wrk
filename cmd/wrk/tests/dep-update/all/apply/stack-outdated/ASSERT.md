## Expected Output

```
==== dep-update ====

  checkout  .
    module  example.com/app
      pin  example.com/lib  v1.0.0 -> v1.2.3
      go mod tidy ok(?:  \(local git\))?
  checkout  external/kool
    module  example.com/kool
      pin  example.com/lib  v1.0.0 -> v1.2.3
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 2, already 0, skipped 0 in 2 checkouts
```

## Expected

- Exit 0.
- `--all` tree with **no** argv `dep` header list.
- Inventory-owned require on the **other** stack checkout is pinned + tidied.
- Summary `updated 2, already 0, skipped 0 in 2 checkouts`.

## Side Effects

- Both consumers require lib@v1.2.3; both have go.sum. Owner go.mod unchanged.

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
      go mod tidy ok(?:  \(local git\))?
  checkout  external/kool
    module  example\.com/kool
      pin  example\.com/lib  v1\.0\.0 -> v1\.2\.3
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 2, already 0, skipped 0 in 2 checkouts
`)
	assertRequireVersion(t, req.ConsumerGoMod, modLib, req.WantVersion)
	assertRequireVersion(t, req.Consumer2GoMod, modLib, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertGoSumExists(t, req.Consumer2ModDir)
	assertOwnerGoModUnchanged(t, req)
}
```
