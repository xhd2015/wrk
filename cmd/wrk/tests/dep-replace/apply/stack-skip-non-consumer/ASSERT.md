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
- Only the gated primary is rewritten.
- Default quiet: no `module  example.com/kool`; no `checkout  external/kool`.
- Other-checkout go.mod identical to baseline.

## Side Effects

- Cross-checkout non-consumer is not mutated and is not listed.

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
	assertNotContains(t, resp.Stdout, "module  "+req.WantConsumer2Module)
	assertNotContains(t, resp.Stdout, "checkout  "+req.WantCheckout2)
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	assertNoTidyArtifacts(t, req)
}
```
