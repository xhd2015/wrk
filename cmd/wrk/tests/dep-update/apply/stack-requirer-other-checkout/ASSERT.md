## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok
  checkout  external/kool
    module  example.com/kool
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok

dep-update: updated 2 modules in 2 checkouts
```

## Expected

- Exit 0.
- Primary **and** other git checkout that already requires xxx are pinned + tidied.
- Checkout labels `.` and `external/kool`.
- Summary `updated 2 modules in 2 checkouts`.

## Side Effects

- Both consumers require @ latest; both have go.sum. kool replace on primary retained.

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
	assert.Output(t, resp.Stdout, `---
version: 3
---
==== dep-update ====
dep  example\.com/dep -> v0\.0\.2(?:  \(tag .+\))?

  checkout  \.
    module  example\.com/app
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok
  checkout  external/kool
    module  example\.com/kool
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok

dep-update: updated 2 modules in 2 checkouts
`)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.Consumer2GoMod, modDep, req.WantVersion)
	assertReplacePresentFor(t, req.ConsumerGoMod, modKool)
	assertGoSumExists(t, req.ConsumerModDir)
	assertGoSumExists(t, req.Consumer2ModDir)
}
```
