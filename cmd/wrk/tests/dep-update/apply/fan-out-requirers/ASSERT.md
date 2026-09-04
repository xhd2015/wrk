## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/app
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok(?:  \(local git\))?
    module  example.com/app/pkg
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 2 modules in 1 checkouts
```

## Expected

- Exit 0.
- Same checkout `.`; root then `pkg/` modules; each pin then tidy.
- Both go.mods: replace dropped; require `example.com/dep` @ `v0.0.2`.
- Both consumers have go.sum.

## Side Effects

- Two existing requirers pinned and tidied; no new requires added elsewhere.

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
      go mod tidy ok(?:  \(local git\))?
    module  example\.com/app/pkg
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 2 modules in 1 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertNoReplaceFor(t, req.Consumer2GoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertRequireVersion(t, req.Consumer2GoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertGoSumExists(t, req.Consumer2ModDir)
}
```
