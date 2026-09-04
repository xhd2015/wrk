## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Only the primary requirer is pinned + tidied.
- Default quiet: no `module  example.com/kool`; no `no require` skip line.
- Other-checkout go.mod identical to baseline.

## Side Effects

- Cross-checkout non-requirer is not mutated and is not listed.

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
    module  example\.com/app
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "module  "+req.WantConsumer2Module)
	assertNotContains(t, resp.Stdout, "no require")
	assertNotContains(t, resp.Stdout, "checkout  "+req.WantCheckout2)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertGoSumExists(t, req.ConsumerModDir)
}
```
