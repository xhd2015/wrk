## Expected Output

```
==== dep-replace --undo ====

  checkout  .
    module  example.com/app
      drop  example.com/dep => <abs>
      skip tidy  (vendor/)

dep-replace: undid 1 replaces in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Banner `--undo`; `drop` line for introduced dep; vendor skips tidy.
- go.mod no longer has replace for `example.com/dep`.
- `require example.com/dep v0.0.1` unchanged.
- No whole-file checkout from HEAD (only surgical dropreplace).

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
	assertUndoBanner(t, resp.Stdout, false)
	assertDropLine(t, resp.Stdout, modDep, false)
	assertContains(t, resp.Stdout, "skip tidy  (vendor/)")
	assertUndoSummary(t, resp.Stdout, 1, 1, 1, false)
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
---
==== dep-replace --undo ====

  checkout  \.
    module  example\.com/app
      drop  example\.com/dep => __ABS__
      skip tidy  \(vendor/\)

dep-replace: undid 1 replaces in 1 modules in 1 checkouts
`)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertRequireVersion(t, req.ConsumerGoMod, modDep, "v0.0.1")
}
```
