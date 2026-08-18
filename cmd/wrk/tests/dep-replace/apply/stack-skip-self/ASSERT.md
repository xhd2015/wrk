## Expected Output

```
==== dep-replace ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/app
      replace  example.com/dep => <abs>

dep-replace: replaced in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Primary consumer rewritten (absolute replace).
- Dep’s own go.mod unchanged (self never rewritten).
- Default quiet: no `module  example.com/dep` consumer block.

## Side Effects

- Only the gated requirer is mutated.

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

  checkout  \.
    module  example\.com/app
      replace  example\.com/dep => __ABS__

dep-replace: replaced in 1 modules in 1 checkouts
`)
	assertNotContains(t, resp.Stdout, "module  "+modDep)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertNoTidyArtifacts(t, req)
}
```
