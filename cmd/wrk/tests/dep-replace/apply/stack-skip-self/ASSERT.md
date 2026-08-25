## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

dep-replace: replaced in 0 modules in 0 checkouts
```

## Expected

- Exit 0.
- Primary keeps relative `replace => ./external/dep` (already equivalent to absDir).
- Dep’s own go.mod unchanged (self never rewritten).
- No consumer module blocks (nothing to write).

## Side Effects

- No go.mod mutation.

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
__ABS__: type=string
---
==== dep-replace ====
dep  example\.com/dep => __ABS__

dep-replace: replaced in 0 modules in 0 checkouts
`)
	assertNotContains(t, resp.Stdout, "module  "+modApp)
	assertNotContains(t, resp.Stdout, "module  "+modDep)
	assertGoModUnchanged(t, req)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
}
```
