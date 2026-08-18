## Expected

- Exit 0.
- Apply tree: banner, `dep` header, checkout `.`, pin, `go mod tidy ok`.
- go.sum exists; require @ latest.
- Wrapper at `go1.19.13` ran (recorded GOROOT/PATH0 contains that pin).

## Side Effects

- Tidy invoked via withgo pin `go1.19.13`, not the host default SDK.

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
      go mod tidy ok

dep-update: updated 1 modules in 1 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, req.WantVersion)
	assertGoSumExists(t, req.ConsumerModDir)
	assertVersionedGoUsed(t, req)
}
```
