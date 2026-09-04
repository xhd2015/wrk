## Expected Output

```
==== dep-update ====
dep  example.com/dep -> v0.0.2

  checkout  .
    module  example.com/consumer
      pin  example.com/dep  v0.0.1 -> v0.0.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Apply banner + argv `dep` header; checkout `.`; pin old -> new; tidy ok.
- Optional `(tag …)` on the `dep` header is allowed.
- No `would:` vocabulary.
- go.mod: no replace for example.com/dep; require at v0.0.2.
- go.sum exists after tidy.

## Side Effects

- Replace dropped; require pinned to latest tag version; tidy ran.

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
    module  example\.com/consumer
      pin  example\.com/dep  v0\.0\.1 -> v0\.0\.2
      go mod tidy ok(?:  \(local git\))?

dep-update: updated 1 modules in 1 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
}
```
